package discover

import (
	"context"
	"sync"
)

// ejcuta el proceso de descubrimiento sobre targets
func Run(ctx context.Context, targets []string, policy Policy) ([]string, error) {
	if !policy.Enabled {
		return targets, nil //si esta deshabilitado, asumimos todos vivos
	}

	aliveHosts := make([]string, 0)
	var mutex sync.Mutex
	var wg sync.WaitGroup

	//canales
	jobs := make(chan string, len(targets))

	//concurrencia
	concurrency := 50
	if len(targets) < concurrency {
		concurrency = len(targets)
	}

	//workers (gorutinas)
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for target := range jobs {
				alive := false

				//probar metodos en orden
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

					//ejecutar discovery
					res, err := d.Discover(ctx, target)
					if err == nil && res.Alive {
						alive = true
						break // si funciona el metodo, el host esta vivo
					}
				}

				if alive {
					mutex.Lock()
					aliveHosts = append(aliveHosts, target)
					mutex.Unlock()
				}
			}
		}()
	}

	//Enviar targets
	for _, t := range targets {
		jobs <- t
	}
	close(jobs)

	wg.Wait()
	return aliveHosts, nil
}
