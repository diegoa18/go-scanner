package core

import (
	"context"
	"fmt"
	"go-scanner/internal/discover/confidence"
	"go-scanner/internal/discover/methods/icmp"
	"go-scanner/internal/discover/methods/tcp"
	"go-scanner/internal/discover/policy"
	"go-scanner/internal/model"
	"sync"
	"time"
)

func Run(ctx context.Context, targets []string, pol policy.Policy) ([]model.HostResult, error) {
	if !pol.Enabled {
		results := make([]model.HostResult, len(targets))
		for i, t := range targets {
			results[i] = model.HostResult{IP: t, Alive: true, Method: "skipped", Reason: "policy-disabled"}
		}
		return results, nil
	}

	if pol.MaxHosts > 0 && len(targets) > pol.MaxHosts {
		return nil, fmt.Errorf("number of targets (%d) exceeds policy limit (%d)", len(targets), pol.MaxHosts)
	}

	var results []model.HostResult
	var mutex sync.Mutex
	var wg sync.WaitGroup

	concurrency := pol.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	if len(targets) < concurrency {
		concurrency = len(targets)
	}

	jobs := make(chan string, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for target := range jobs {
				if pol.Delay > 0 {
					select {
					case <-time.After(pol.Delay):
					case <-ctx.Done():
						return
					}
				}

				var signals []confidence.ResultSignal
				var bestResult model.HostResult
				bestResult.IP = target

				for _, method := range pol.Methods {
					var discoverer Discoverer
					switch method {
					case "icmp":
						discoverer = icmp.NewDiscoverer(pol.Timeout)
					case "tcp-connect":
						discoverer = tcp.NewConnectDiscoverer([]int{80, 443}, pol.Timeout)
					default:
						continue
					}

					res, err := discoverer.Discover(ctx, target)

					signals = append(signals, confidence.ResultSignal{
						Method: method,
						Alive:  res.Alive,
						Reason: res.Reason,
						Error:  err,
					})

					if err != nil {
						if bestResult.Reason == "" {
							bestResult.Error = err
							bestResult.Reason = res.Reason
						}
						continue
					}

					if res.Alive {
						if !bestResult.Alive || res.RTT < bestResult.RTT {
							bestResult = res
						}
					} else {
						if !bestResult.Alive && bestResult.Reason == "" {
							bestResult = res
						}
					}
				}

				calc := confidence.Calculate(signals)
				bestResult.Confidence = calc.Level
				bestResult.Score = calc.Score

				if bestResult.Alive {
					bestResult.Reason = fmt.Sprintf("%s | confidence: %s", bestResult.Reason, calc.Reason)
					mutex.Lock()
					results = append(results, bestResult)
					mutex.Unlock()
				}
			}
		}()
	}

	for _, t := range targets {
		jobs <- t
	}
	close(jobs)

	wg.Wait()
	return results, nil
}
