package main

import (
	"fmt"
	"go-scanner/internal/config"
	"go-scanner/internal/report"
	"go-scanner/internal/scanner"
	"os"
	"time"
)

func main() {
	//cargar la configuracion
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("starting scan to %s (%d ports)\n", cfg.Target, len(cfg.Ports))
	fmt.Printf("concurrency: %d workers | timeout: %v\n", cfg.Concurrency, cfg.Timeout)

	//instanciar scanner
	tcpScanner := scanner.NewTCPConnectScanner(
		cfg.Target,
		cfg.Ports,
		cfg.Timeout,
		cfg.Concurrency,
	)

	//preparar canal de resultados
	results := make(chan scanner.ScanResult)

	//ejecutar scanner
	startTime := time.Now()
	go tcpScanner.Scan(results)

	//procesar y reportar resultados
	report.PrintResults(results)

	elapsed := time.Since(startTime)
	fmt.Printf("scan finished in %v\n", elapsed)
}
