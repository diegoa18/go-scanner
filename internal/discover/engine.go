package discover

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ejecuta el proceso de descubrimiento sobre targets
func Run(ctx context.Context, targets []string, policy Policy) ([]HostResult, error) {
	if !policy.Enabled {
		//si el discovery esta deshabilitado, retornamos targets como si estuvieran vivos
		results := make([]HostResult, len(targets))
		for i, t := range targets {
			results[i] = HostResult{IP: t, Alive: true, Method: "skipped", Reason: "policy-disabled"}
		}
		return results, nil
	}

	//validar maxhosts
	if policy.MaxHosts > 0 && len(targets) > policy.MaxHosts {
		return nil, fmt.Errorf("number of targets (%d) exceeds policy limit (%d)", len(targets), policy.MaxHosts)
	}

	var results []HostResult
	var mutex sync.Mutex
	var wg sync.WaitGroup

	//worker pool
	concurrency := policy.Concurrency
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
				if policy.Delay > 0 {
					select {
					case <-time.After(policy.Delay):
					case <-ctx.Done():
						return
					}
				}

				//logica de escaneo
				bestResult := HostResult{IP: target, Alive: false}

				for _, method := range policy.Methods {
					var d Discoverer
					switch method {
					case "icmp":
						d = NewICMPDiscoverer(policy.Timeout)
					case "tcp-connect":
						d = NewTCPConnectDiscoverer([]int{80, 443}, policy.Timeout)
					default:
						continue
					}

					res, err := d.Discover(ctx, target)

					//en caso de error, registramos pero seguimos intentando otros metodos
					if err != nil {
						//el primer error se guarda
						if bestResult.Error == nil {
							bestResult.Error = err
							bestResult.Reason = res.Reason
						}
						continue
					}

					if res.Alive {
						bestResult = res
						break //si encuentra un host vivo, sale del bucle
					}

					//si no esta vivo, actualizamos la razón
					bestResult.Reason = res.Reason
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
