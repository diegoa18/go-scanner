package cli

import (
	"context"
	"flag"
	"fmt"
	"go-scanner/internal/model"
	"go-scanner/internal/orchestrator"
	"go-scanner/internal/profile"
	"go-scanner/internal/report"
	"go-scanner/internal/scanner"
	"go-scanner/internal/scanner/tcp"
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
		fmt.Println("Error: target required (IP, CIDR or Range)")
		fmt.Println("Usage: go-scanner tcp connect -p <ports> <target>")
		os.Exit(1)
	}
	rawTarget := cmd.Arg(0)

	//parseo de targets (CIDR, Range, Single)
	targets, err := utils.ParseTarget(rawTarget)
	if err != nil {
		fmt.Printf("Error validating target: %v\n", err)
		os.Exit(1)
	}

	if len(targets) == 0 {
		fmt.Println("Error: No valid targets found.")
		os.Exit(1)
	}

	//parseo de puertos
	ports, err := utils.ParsePortRange(*portRange)
	if err != nil {
		fmt.Printf("Error parsing ports: %v\n", err)
		os.Exit(1)
	}

	//cargar perfil
	selectedProfile, ok := profile.Get(*profileName)
	if !ok {
		fmt.Printf("Error: unknown profile '%s'\n", *profileName)
		fmt.Printf("Available profiles: %v\n", profile.Available())
		os.Exit(1)
	}

	//configurar policy
	policy := selectedProfile.Policy
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

	if *probeFlag {
		policy.ActiveProbing = true
		policy.AllowedProbes = activeProbes
	}

	//resumen
	fmt.Printf("Profile: %s (%s)\n", selectedProfile.Name, selectedProfile.Description)
	fmt.Printf("Targets: %d | Ports: %d | Workers: %d | Timeout: %v\n",
		len(targets), len(ports), policy.Concurrency, policy.Timeout)

	if policy.Discovery.Enabled {
		fmt.Printf("Discovery Enabled: %v (Timeout: %v)\n", policy.Discovery.Methods, policy.Discovery.Timeout)
	}

	//SCANNER FACTORY
	factory := func(t string, meta *model.HostMetadata) scanner.Scanner {
		//crea scanner para este target especifico
		return tcp.NewTCPConnectScanner(
			t,
			ports,
			policy.Timeout,
			policy.Concurrency,
			*banner,
			meta,
		)
	}

	//ORQUESTACION

	//coordinador
	coord := orchestrator.NewCoordinator(policy, factory)

	ctx := context.Background()
	startTime := time.Now()

	//ejecutar
	resultsChan := coord.Run(ctx, targets)
	report.PrintResults(resultsChan)

	elapsed := time.Since(startTime)
	fmt.Printf("Campaign completed in %v\n", elapsed)
}
