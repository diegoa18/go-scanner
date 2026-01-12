package scan

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go-scanner/internal/config/profile"
	"go-scanner/internal/model"
	"go-scanner/internal/orchestrator"
	"go-scanner/internal/scanner"
	"go-scanner/internal/scanner/tcp"
	"go-scanner/internal/utils"

	"github.com/google/uuid"
)

// define contrato para la capa de aplicacioon de escaneo
type Service interface {
	Run(ctx context.Context, req ScanRequest) (*ScanReport, error)
}

// interfaz service
type service struct {
}

// nueva instancia
func NewService() Service {
	return &service{}
}

// ejecuta un escaneo síncrono basado en el request
func (s *service) Run(ctx context.Context, req ScanRequest) (*ScanReport, error) {
	jobID := uuid.New().String()
	startTime := time.Now()

	// validar inputs básicos
	if len(req.Targets) == 0 {
		return nil, errors.New("no targets specified")
	}

	// resolver perfil
	profileName := req.ProfileName
	if profileName == "" {
		profileName = "default"
	}

	selectedProfile, ok := profile.Get(profileName)
	if !ok {
		return nil, fmt.Errorf("unknown profile: %s", profileName)
	}

	// configurar policy (mezclando perfil y opciones del request)
	policy := selectedProfile.Policy
	s.applyOptions(&policy, req.Options)

	// parsear puertos
	portStr := req.Ports
	if portStr == "" {
		portStr = "80,443,22,21,25,8080,8443,3306,5432,3389" //puertos predefinidos
	}
	ports, err := utils.ParsePortRange(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid ports: %w", err)
	}

	// normalizar targets
	var finalTargets []string
	for _, t := range req.Targets {
		expanded, err := utils.ParseTarget(t)
		if err != nil {
			return nil, fmt.Errorf("invalid target '%s': %w", t, err)
		}
		finalTargets = append(finalTargets, expanded...)
	}

	if len(finalTargets) == 0 {
		return nil, errors.New("no valid targets found after parsing")
	}

	// preparar factory de scanners
	factory := func(t string, meta *model.HostMetadata) scanner.Scanner {
		return tcp.NewTCPConnectScanner(
			t,
			ports,
			policy.Timeout,
			policy.Concurrency,
			req.Options.Banner, //banner override explicito o false
			meta,
		)
	}

	// RUN (orquestacion)
	coord := orchestrator.NewCoordinator(policy, factory)
	resultsChan, errChan := coord.Run(ctx, finalTargets)

	// estado inicial
	currentStatus := StatusRunning
	var scanErrors []ScanError
	var scanResults []scanner.ScanResult

	//loop de recoleccion, se necesita que ambos canales cierren
	for resultsChan != nil || errChan != nil {
		select {
		case res, ok := <-resultsChan:
			if !ok {
				resultsChan = nil
				continue
			}
			scanResults = append(scanResults, res)

		case err, ok := <-errChan:
			if !ok {
				errChan = nil
				continue
			}
			// capturar error
			scanErrors = append(scanErrors, NewScanError("orchestrator", err, ""))
		}
	}

	// determinacion del estado final
	if len(scanErrors) > 0 {
		if len(scanResults) > 0 {
			currentStatus = StatusPartial
		} else {
			currentStatus = StatusFailed
		}

	} else {
		if len(scanResults) > 0 {
			currentStatus = StatusCompleted
		} else {

			//PENDIENTE -> ACTUALMENTE COMPLETED SI NO HUBO ERRORES EXPLICITOS, CONSIDERAR ANOMALIAS
			currentStatus = StatusCompleted
		}
	}

	// construir reporte
	elapsed := time.Since(startTime)

	return &ScanReport{
		JobID:   jobID,
		Status:  currentStatus,
		Results: scanResults,
		Errors:  scanErrors,
		Metadata: ExecutionMetadata{
			Duration:    elapsed,
			TargetCount: len(finalTargets),
			ProfileUsed: selectedProfile.Name,
		},
	}, nil
}

// aplicar sobreescrituras a la politica base.
func (s *service) applyOptions(p *orchestrator.ScanPolicy, opts ScanOptions) {
	if opts.TimeoutMs > 0 {
		p.Timeout = time.Duration(opts.TimeoutMs) * time.Millisecond
	}
	if opts.Concurrency > 0 {
		p.Concurrency = opts.Concurrency
	}
	// aplicar configuracion de probes activos
	if opts.Probe {
		p.ActiveProbing = true
		if len(opts.ProbeTypes) > 0 {
			p.AllowedProbes = opts.ProbeTypes
		}
		// si el probe -> true, utilizamos http/https por default
		if len(p.AllowedProbes) == 0 {
			p.AllowedProbes = []string{"http", "https"}
		}
	}
}
