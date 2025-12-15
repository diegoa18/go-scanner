package cli

import (
	"context"
	"flag"
	"fmt"
	"go-scanner/internal/config"
	"go-scanner/internal/orchestrator"
	"go-scanner/internal/profile"
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
	profileName := cmd.String("profile", "default", "Scan profile: passive, default, aggressive")
	portRange := cmd.String("p", "1-1024", "Ports to scan (e.g: '80', '1-1024', '80,443')")
	timeoutMs := cmd.Int("timeout", -1, "Timeout per connection in ms (default: from profile)")
	concurrency := cmd.Int("threads", -1, "Maximum number of concurrent connections (default: from profile)")
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

	//cargar perfil base
	selectedProfile, ok := profile.Get(*profileName)
	if !ok {
		fmt.Printf("Error: unknown profile '%s'\n", *profileName)
		fmt.Printf("Available profiles: %v\n", profile.Available())
		os.Exit(1)
	}

	//construir policy desde perfil
	policy := selectedProfile.Policy

	//aplicar overrides explicitos de las flags
	if *timeoutMs > 0 {
		policy.Timeout = time.Duration(*timeoutMs) * time.Millisecond
	}
	if *concurrency > 0 {
		policy.Concurrency = *concurrency
	}

	//parsear probes types
	activeProbes := strings.Split(*probeTypes, ",")
	for i := range activeProbes {
		activeProbes[i] = strings.TrimSpace(strings.ToLower(activeProbes[i]))
	}

	//override de active probing si --probe fue escrito
	if *probeFlag {
		policy.ActiveProbing = true
		policy.AllowedProbes = activeProbes
	}

	//crear configuracion (solo para scanner base)
	cfg := &config.Config{
		Target:       resolvedIP,
		PortRange:    *portRange,
		Ports:        ports,
		Timeout:      policy.Timeout,
		Concurrency:  policy.Concurrency,
		EnableBanner: *banner,
		EnableProbe:  policy.ActiveProbing,
		ProbeTypes:   policy.AllowedProbes,
	}

	if err := cfg.Validate(); err != nil {
		fmt.Printf("Configuration error: %v\n", err)
		os.Exit(1)
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

	//mostrar configuración
	fmt.Printf("Profile: %s (%s)\n", selectedProfile.Name, selectedProfile.Description)
	fmt.Printf("Starting TCP Connect scan to %s (Range: %s, %d ports)\n",
		cfg.Target,
		cfg.PortRange,
		len(cfg.Ports))

	fmt.Printf("Configuration: %d workers | Timeout: %v\n",
		policy.Concurrency,
		policy.Timeout)

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
