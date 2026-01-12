package orchestrator

import (
	"context"
	"fmt"
	"go-scanner/internal/discover"
	"go-scanner/internal/model"
	"go-scanner/internal/scanner"
)

// define una funcion que crea un scanner para un target dado
type ScannerFactory func(target string, meta *model.HostMetadata) scanner.Scanner

// orquesta la ejecucion sobre multiples targets
type Coordinator struct {
	Policy  ScanPolicy
	Factory ScannerFactory
}

func NewCoordinator(policy ScanPolicy, factory ScannerFactory) *Coordinator {
	return &Coordinator{
		Policy:  policy,
		Factory: factory,
	}
}

// ejecuta descubrimiento y luego escaneo para cada host
// uni-canal de reusltados y errores
func (c *Coordinator) Run(ctx context.Context, targets []string) (<-chan scanner.ScanResult, <-chan error) {
	out := make(chan scanner.ScanResult)
	errChan := make(chan error, 1) // buffered para no bloquear si main no lee inmediatamente

	go func() {
		defer close(out)
		defer close(errChan)

		//mapa de metadatos
		metadataMap := make(map[string]*model.HostMetadata)

		//fase de descubrimiento
		scannableTargets := targets
		if c.Policy.Discovery.Enabled {
			fmt.Printf("Starting discovery phase on %d targets...\n", len(targets))
			aliveResults, err := discover.Run(ctx, targets, c.Policy.Discovery)

			if err != nil {
				// Propagar error crítico de discovery
				errChan <- fmt.Errorf("discovery phase critical failure: %w", err)
				return // Abortar ejecución
			}

			//extraer IPs de los resultados vivos
			aliveIPs := make([]string, 0, len(aliveResults))
			for _, r := range aliveResults {
				if r.Alive {
					aliveIPs = append(aliveIPs, r.IP)
					//popular metadatos
					metadataMap[r.IP] = &model.HostMetadata{
						ID:              r.IP,
						DiscoveryMethod: r.Method,
						DiscoveryRTT:    r.RTT,
						DiscoveryReason: r.Reason,
						DiscoveryTime:   r.Timestamp,
						Confidence:      r.Confidence, // Usamos el calculado
					}
				}
			}
			scannableTargets = aliveIPs
			fmt.Printf("Discovery complete. %d/%d hosts alive.\n", len(scannableTargets), len(targets))
		}

		//iterar sobre targets
		for _, target := range scannableTargets {
			//revisar contexto
			select {
			case <-ctx.Done():
				return
			default:
			}

			meta := metadataMap[target]

			//en caso de si discovery esta deshabilitado, hace que meta sea nil pues
			if meta == nil {
				meta = &model.HostMetadata{
					ID:         target,
					Confidence: "unknown",
				}
			}

			//crear scanner via factory
			s := c.Factory(target, meta)

			//crear engine
			engine := NewEngine(c.Policy, target, nil, s)

			//ejecutar engine
			// PENDIENTE -> el engine debe retornar errores en caso de fallo
			results := engine.Run(ctx)

			//resultados
			for res := range results {
				select {
				case <-ctx.Done():
					return
				case out <- res:
				}
			}
		}
	}()

	return out, errChan
}
