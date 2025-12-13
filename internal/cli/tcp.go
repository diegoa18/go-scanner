package cli

import (
	"flag"
	"fmt"
	"go-scanner/internal/config"
	"go-scanner/internal/report"
	"go-scanner/internal/scanner" //import service detection
	"go-scanner/internal/service"
	"go-scanner/internal/utils"
	"os"
	"time"
)

// maneja el comando top "tcp" y sus subcomandos
func handleTCPCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: go-scanner tcp <subcommand> [options] <target>")
		fmt.Println("Subcommands available:")
		fmt.Println("  connect    Perform a complete TCP Connect scan")
		os.Exit(1)
	}

	subcommand := args[0]
	switch subcommand {
	case "connect":
		handleTCPConnect(args[1:])
	default:
		fmt.Printf("Subcommand unknown for tcp: %s\n", subcommand)
		os.Exit(1)
	}
}

// logica especifica para "tcp connect"
func handleTCPConnect(args []string) {
	//flagset propio, en caso de conflictos
	cmd := flag.NewFlagSet("connect", flag.ExitOnError)

	//flags de tcp connect
	portRange := cmd.String("p", "1-1024", "Ports to scan (e.g: '80', '1-1024', '80,443')")
	timeoutMs := cmd.Int("timeout", 1000, "Timeout per connection in ms")
	concurrency := cmd.Int("threads", 100, "Maximum number of concurrent connections")
	banner := cmd.Bool("banner", false, "Enable passive banner grabbing on supported ports")

	cmd.Parse(args) //parsear flags

	//validar argumentos
	if cmd.NArg() < 1 {
		fmt.Println("Error: target required (IP or Host)")
		fmt.Println("Usage: go-scanner tcp connect -p <ports> <target>")
		os.Exit(1)
	}
	target := cmd.Arg(0)

	//utilizar resolucion de IP
	resolvedIP, err := utils.Resolve(target)
	if err != nil {
		fmt.Printf("Error resolving target: %v\n", err)
		os.Exit(1)
	}

	if resolvedIP != target {
		fmt.Printf("Resolved %s to %s\n", target, resolvedIP)
	}

	//parsear puertos con internal/utils
	ports, err := utils.ParsePortRange(*portRange)
	if err != nil {
		fmt.Printf("Error parsing ports: %v\n", err)
		os.Exit(1)
	}

	//crear configuracion
	cfg := &config.Config{
		Target:       resolvedIP, //utilizar IP resuelta
		PortRange:    *portRange,
		Ports:        ports,
		Timeout:      time.Duration(*timeoutMs) * time.Millisecond,
		Concurrency:  *concurrency,
		EnableBanner: *banner, //banner grabbing
	}

	//ejecutar escaneo
	fmt.Printf("Starting TCP Connect scan to %s (Range: %s, %d ports)\n", cfg.Target, cfg.PortRange, len(cfg.Ports))
	fmt.Printf("Configuration: %d workers | Timeout: %v\n", cfg.Concurrency, cfg.Timeout)

	//instanciar scanner
	tcpScanner := scanner.NewTCPConnectScanner(
		cfg.Target,
		cfg.Ports,
		cfg.Timeout,
		cfg.Concurrency,
		cfg.EnableBanner,
	)

	//preparar canal de resultados
	results := make(chan scanner.ScanResult)

	//tiempo
	startTime := time.Now()
	go tcpScanner.Scan(results)

	//resultados enriquecidos usando service detection
	enrichedResults := make(chan scanner.ScanResult)

	go func() {
		defer close(enrichedResults)

		for res := range results {
			if res.IsOpen {
				//aqui es donde se detecta el servicio, no en el core del scanner
				svcInfo := service.Detect(res.Port, res.Banner)
				res.Service = string(svcInfo.Type) //asignacion del tipo detectado
			}
			enrichedResults <- res
		}
	}()

	//ahora si, procesar los resultados
	report.PrintResults(enrichedResults)

	elapsed := time.Since(startTime)
	fmt.Printf("Scan completed in %v\n", elapsed)
}
