package core

import (
	"context"
	"fmt"
	"go-scanner/internal/discover/confidence"
	"go-scanner/internal/discover/methods/icmp"
	"go-scanner/internal/discover/methods/tcp"
	"go-scanner/internal/discover/policy"
	"sync"
	"time"
)

// ejecuta el proceso de descubrimiento sobre targets
func Run(ctx context.Context, targets []string, pol policy.Policy) ([]HostResult, error) {
	if !pol.Enabled {
		//si el discovery esta deshabilitado, retornamos targets como si estuvieran vivos
		results := make([]HostResult, len(targets))
		for i, t := range targets {
			results[i] = HostResult{IP: t, Alive: true, Method: "skipped", Reason: "policy-disabled"}
		}
		return results, nil
	}

	//validar maxhosts
	if pol.MaxHosts > 0 && len(targets) > pol.MaxHosts {
		return nil, fmt.Errorf("number of targets (%d) exceeds policy limit (%d)", len(targets), pol.MaxHosts)
	}

	var results []HostResult
	var mutex sync.Mutex
	var wg sync.WaitGroup

	//worker pool
	concurrency := pol.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	if len(targets) < concurrency {
		concurrency = len(targets)
	}

	jobs := make(chan string, concurrency)

	//workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for target := range jobs {
				// Delay opcional entre targets por worker
				if pol.Delay > 0 {
					select {
					case <-time.After(pol.Delay):
					case <-ctx.Done():
						return
					}
				}

				//logica de escaneo
				var signals []confidence.ResultSignal
				var bestResult HostResult
				bestResult.IP = target

				// Ejecutar TODOS los metodos para tener multiples senales
				for _, method := range pol.Methods {
					switch method {
					case "icmp":
						d := icmp.NewDiscoverer(pol.Timeout)
						res, err := d.Discover(ctx, target)

						// convertimos de icmp.HostResult a core.HostResult
						coreRes := HostResult{
							IP:         res.IP,
							Alive:      res.Alive,
							RTT:        res.RTT,
							Method:     res.Method,
							Reason:     res.Reason,
							Error:      res.Error,
							Timestamp:  res.Timestamp,
							Confidence: res.Confidence,
							Score:      res.Score,
						}

						// Registrar señal para el confidence engine
						sig := confidence.ResultSignal{
							Method: method,
							Alive:  coreRes.Alive,
							Reason: coreRes.Reason,
							Error:  err,
						}
						signals = append(signals, sig)

						if err != nil {
							if bestResult.Reason == "" {
								bestResult.Error = err
								bestResult.Reason = coreRes.Reason
							}
							continue
						}

						if coreRes.Alive {
							if !bestResult.Alive || coreRes.RTT < bestResult.RTT {
								bestResult = coreRes
							}
						} else {
							if !bestResult.Alive && bestResult.Reason == "" {
								bestResult = coreRes
							}
						}

					case "tcp-connect":
						d := tcp.NewConnectDiscoverer([]int{80, 443}, pol.Timeout)
						res, err := d.Discover(ctx, target)

						// convertir de tcp.HostResult a core.HostResult
						coreRes := HostResult{
							IP:         res.IP,
							Alive:      res.Alive,
							RTT:        res.RTT,
							Method:     res.Method,
							Reason:     res.Reason,
							Error:      res.Error,
							Timestamp:  res.Timestamp,
							Confidence: res.Confidence,
							Score:      res.Score,
						}

						// registrar señal para el confidence engine
						sig := confidence.ResultSignal{
							Method: method,
							Alive:  coreRes.Alive,
							Reason: coreRes.Reason,
							Error:  err,
						}
						signals = append(signals, sig)

						if err != nil {
							if bestResult.Reason == "" {
								bestResult.Error = err
								bestResult.Reason = coreRes.Reason
							}
							continue
						}

						if coreRes.Alive {
							if !bestResult.Alive || coreRes.RTT < bestResult.RTT {
								bestResult = coreRes
							}
						} else {
							if !bestResult.Alive && bestResult.Reason == "" {
								bestResult = coreRes
							}
						}

					default:
						continue
					}
				}

				// calcular confianza
				calc := confidence.Calculate(signals)
				bestResult.Confidence = calc.Level
				bestResult.Score = calc.Score

				// actualizar razon compuesta si se desea, o dejar la del metodo principal
				// agregamos la razon de confianza a la razon de descubrimiento para trazabilidad
				if bestResult.Alive {
					bestResult.Reason = fmt.Sprintf("%s | confidence: %s", bestResult.Reason, calc.Reason)
				} else if calc.Level != "unknown" && calc.Level != "low" {
					//PENDIENTE: evaluar si es necesario
				}

				if bestResult.Alive {
					mutex.Lock()
					results = append(results, bestResult)
					mutex.Unlock()
				}
			}
		}()
	}

	//enviar targets
	for _, t := range targets {
		jobs <- t
	}
	close(jobs)

	wg.Wait()
	return results, nil
}
