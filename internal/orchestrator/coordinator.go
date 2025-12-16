package orchestrator

import (
	"context"
	"fmt"
	"go-scanner/internal/discover"
	"go-scanner/internal/scanner"
	// Faltaba importar sync? No, Engine lo usaba.
)

// ScannerFactory define una funcion que crea un scanner para un target dado
type ScannerFactory func(target string) scanner.Scanner

// Coordinator orquesta la ejecucion sobre multiples targets
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

// Run ejecuta descubrimiento (si aplica) y luego escaneo para cada host
// Retorna un canal unificado de resultados
func (c *Coordinator) Run(ctx context.Context, targets []string) <-chan scanner.ScanResult {
	out := make(chan scanner.ScanResult)

	go func() {
		defer close(out)

		// 1. Discovery Phase
		scannableTargets := targets
		if c.Policy.Discovery.Enabled {
			fmt.Printf("Starting discovery phase on %d targets...\n", len(targets))
			alive, err := discover.Run(ctx, targets, c.Policy.Discovery)
			if err != nil {
				// Log error?
				fmt.Printf("Discovery error: %v\n", err)
				return
			}
			scannableTargets = alive
			fmt.Printf("Discovery complete. %d/%d hosts alive.\n", len(alive), len(targets))
		}

		// 2. Access Scannable Targets
		// Si es muchos hosts, queremos paralelizar el escaneo DE HOSTS o solo los puertos?
		// Engine paralela puertos. Si corremos 100 engines a la vez, explotamos.
		// Deberimos tener un semaforo para Hosts concurrentes.
		// ScanPolicy.Concurrency es "concurrencia total del sistema" o "por host"?
		// En Policy dice "Maximum number of concurrent connections".
		// Si es por host, hay que serializar hosts o dividir concurrencia.
		// Por simplicidad en esta iteracion: Serial o worker pool limitado de hosts.

		// Worker pool para hosts

		// Si policy.Concurrency es 100, y hostConcurrency es 5, cada host usa 20? No, Engine usa Policy.Concurrency.
		// Para no saturar FD, corremos hosts secuenciales o pocos paralelos.
		// Vamos a correr secuencial por ahora para respetar limites de policy global (ya que policy se pasa a cada engine).

		for _, target := range scannableTargets {
			// Check context
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Crear Scanner via Factory
			s := c.Factory(target)

			// Crear Engine
			// Engine usa la Policy Global. Si la policy dice 100 hilos, usara 100 hilos.
			engine := NewEngine(c.Policy, target, nil, s) // Ports? El scanner ya tiene los puertos configurados en el Factory?
			// Engine struct tiene Ports []int. Pero Engine no los usa para escanear, solo para Data Structs?
			// Engine.Run llama a Scanner.Scan(). El Scanner ya sabe sus puertos.
			// Engine.Ports es redundante o informativo?
			// En Engine.Run no se usa e.Ports. Revisar engine.go.
			// Engine struct field Ports is used?
			// NewEngine(..., ports, ...)
			// En internal/orchestrator/engine.go:
			// func NewEngine... targets, ports...
			// e.Ports = ports
			// Pero Run() no usa e.Ports.
			// OK, pasamos nil o vacio si el Factory ya configuro el scanner.
			// El issue es que scanner.Scanner interface Scan() devuelve resultados que tiene el puerto.

			// Run engine
			results := engine.Run(ctx)

			// Forward results
			for res := range results {
				out <- res
			}
		}
	}()

	return out
}
