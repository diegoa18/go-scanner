package cli

import (
	"context"
	"flag"
	"fmt"
	"go-scanner/internal/config"
	"go-scanner/internal/orchestrator"
	"go-scanner/internal/report"
	"go-scanner/internal/scanner"
	"go-scanner/internal/utils"
	"os"
	"strings"
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
	probeFlag := cmd.Bool("probe", false, "Enable ACTIVE probing on detected services")
	probeTypes := cmd.String("probe-types", "http,https", "Comma-separated list of probe types to run (default: http,https)")

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

	//parsear probes types
	activeProbes := strings.Split(*probeTypes, ",")
	for i := range activeProbes {
		activeProbes[i] = strings.TrimSpace(strings.ToLower(activeProbes[i]))
	}

	//crear configuracion
	cfg := &config.Config{
		Target:       resolvedIP, //utilizar IP resuelta
		PortRange:    *portRange,
		Ports:        ports,
		Timeout:      time.Duration(*timeoutMs) * time.Millisecond,
		Concurrency:  *concurrency,
		EnableBanner: *banner,      //banner grabbing
		EnableProbe:  *probeFlag,   //active probing
		ProbeTypes:   activeProbes, //tipos de probes
	}

	if err := cfg.Validate(); err != nil {
		fmt.Printf("Configuration error: %v\n", err)
		os.Exit(1)
	}

	//contruccion de policy
	policy := orchestrator.ScanPolicy{
		Timeout:          cfg.Timeout,
		Concurrency:      cfg.Concurrency,
		ServiceDetection: true,
		ActiveProbing:    cfg.EnableProbe,
		AllowedProbes:    cfg.ProbeTypes,
	}

	//scanner base
	tcpScanner := scanner.NewTCPConnectScanner(
		cfg.Target,
		cfg.Ports,
		cfg.Timeout,
		cfg.Concurrency,
		cfg.EnableBanner,
	)

	//motor de ejecucion
	engine := orchestrator.NewEngine(policy, cfg.Target, cfg.Ports, tcpScanner)
	fmt.Printf("Starting TCP Connect scan to %s (Range: %s, %d ports)\n",
		cfg.Target,
		cfg.PortRange,
		len(cfg.Ports))

	fmt.Printf("Configuration: %d workers | Timeout: %v\n",
		cfg.Concurrency,
		cfg.Timeout)

	if policy.ActiveProbing {
		fmt.Printf("Active Probing ENABLED: %v\n",
			policy.AllowedProbes)
	}

	//ejecutar la pipeline
	ctx := context.Background()
	startTime := time.Now()

	//resultados procesados por el motor
	results := engine.Run(ctx)

	//reportar
	report.PrintResults(results)

	elapsed := time.Since(startTime)
	fmt.Printf("Scan completed in %v\n", elapsed)
}
