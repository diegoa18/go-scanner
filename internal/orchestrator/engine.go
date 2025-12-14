package orchestrator

import (
	"context"
	"go-scanner/internal/probe"
	"go-scanner/internal/scanner"
	"go-scanner/internal/service"
	"strings"
	"sync"
	"time"
)

// pipeline: scanner -> service detection -> active probing
type Engine struct {
	Policy  ScanPolicy
	Target  string
	Ports   []int
	Scanner scanner.Scanner
}

// nueva instancia de pipeline
func NewEngine(policy ScanPolicy, target string, ports []int, baseScanner scanner.Scanner) *Engine {
	return &Engine{
		Policy:  policy,
		Target:  target,
		Ports:   ports,
		Scanner: baseScanner,
	}
}

// ejecuta pipeline y retorna resultados (con gorutinas)
func (e *Engine) Run(ctx context.Context) <-chan scanner.ScanResult {
	//canal de salida
	out := make(chan scanner.ScanResult)

	//canal interno para resultados crudos
	//el scanner base no sabe de contextos aun (PENDIENTE)
	rawResults := make(chan scanner.ScanResult)

	//GORUTINA PRINCIPAL
	go func() {
		defer close(out)

		//escaneo base
		go e.Scanner.Scan(rawResults)

		//stream processor simple
		enrichmentWG := sync.WaitGroup{}

		//PENDIENTE -> en caso de lentitud, implementar worker pool o algo asi lol

		//recorrer resultados crudos
		for res := range rawResults {
			enrichmentWG.Add(1)
			go func(r scanner.ScanResult) {
				defer enrichmentWG.Done()

				//verificar cancelacion de contexto
				select {
				case <-ctx.Done():
					return
				default:
				}

				//procesar resultado
				enriched := e.processResult(r)

				//enviar a la salida
				out <- enriched
			}(res)
		}

		enrichmentWG.Wait()
	}()

	return out
}

// logica de negocio sobre un resultado crudo
func (e *Engine) processResult(res scanner.ScanResult) scanner.ScanResult {
	if !res.IsOpen {
		return res
	}

	//deteccion de servicio (pasiva o activa)
	if e.Policy.ServiceDetection {
		svcInfo := service.Detect(res.Port, res.Banner)
		res.Service = string(svcInfo.Type)

		//probing activo
		if e.Policy.ActiveProbing {
			e.applyActiveProbe(&res, svcInfo.Type)
		}
	}
	return res
}

// aplica probes activos
func (e *Engine) applyActiveProbe(res *scanner.ScanResult, svcType service.ServiceType) {
	serviceName := string(svcType)

	//buscar prober en probe/registry
	prober, found := probe.Get(serviceName)
	if !found {
		return
	}

	//verficar si el prober esta permitido
	allowed := false
	lowerService := strings.ToLower(serviceName)

	if len(e.Policy.AllowedProbes) == 0 {
		//asumimos nada por seguridad, actualmente el CLI deja como default "http,https"
		//falta implementar mas policy
		return
	}

	//iterar sobre la lista de allowedProbes
	for _, t := range e.Policy.AllowedProbes {
		if t == "all" || t == lowerService {
			allowed = true
			break
		}
	}

	if !allowed {
		return
	}

	//timeout minimo de 3s
	probeTimeout := 3 * time.Second

	//puede que en policy defina un timeout global

	probeBanner, err := prober.Probe(e.Target, res.Port, probeTimeout)
	if err == nil && probeBanner != "" {
		if res.Banner != "" {
			res.Banner = res.Banner + " | " + probeBanner
		} else {
			res.Banner = probeBanner
		}
	}
}
