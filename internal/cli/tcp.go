package cli

import (
	"flag"
	"fmt"
	"go-scanner/internal/config"
	"go-scanner/internal/report"
	"go-scanner/internal/scanner"
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

	cmd.Parse(args) //parsear flags

	//validar argumentos
	if cmd.NArg() < 1 {
		fmt.Println("Error: target required (IP or Host)")
		fmt.Println("Usage: go-scanner tcp connect -p <ports> <target>")
		os.Exit(1)
	}
	target := cmd.Arg(0)

	//validar IP con internal/utils
	if !utils.IsValidIP(target) {
		//SE PODRIA AGREGAR UNA RESOLUCION DNS :3
		//por ahora adevertencia con un print penca
		fmt.Printf("warning: %s its seems not a valid IP\n", target)
	}

	//parsear puertos con internal/utils
	ports, err := utils.ParsePortRange(*portRange)
	if err != nil {
		fmt.Printf("Error parsing ports: %v\n", err)
		os.Exit(1)
	}

	//crear configuracion
	cfg := &config.Config{
		Target:      target,
		PortRange:   *portRange,
		Ports:       ports,
		Timeout:     time.Duration(*timeoutMs) * time.Millisecond,
		Concurrency: *concurrency,
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
	)

	//preparar canal de resultados
	results := make(chan scanner.ScanResult)

	//tiempo
	startTime := time.Now()
	go tcpScanner.Scan(results)

	//procesar resultados
	report.PrintResults(results)

	elapsed := time.Since(startTime)
	fmt.Printf("Scan completed in %v\n", elapsed)
}
