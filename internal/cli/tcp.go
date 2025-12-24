package cli

import (
	"context"
	"flag"
	"fmt"
	"go-scanner/internal/app/scan"
	"go-scanner/internal/report"
	"go-scanner/internal/utils"
	"os"
	"strings"
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
	_, err = utils.ParsePortRange(*portRange)
	if err != nil {
		fmt.Printf("Error parsing ports: %v\n", err)
		os.Exit(1)
	}

	//parsear probes types
	activeProbes := strings.Split(*probeTypes, ",")
	for i := range activeProbes {
		activeProbes[i] = strings.TrimSpace(strings.ToLower(activeProbes[i]))
	}

	//configurar request
	req := scan.ScanRequest{
		Targets:     targets,
		Ports:       *portRange,
		ProfileName: *profileName,
		Options: scan.ScanOptions{
			TimeoutMs:   *timeoutMs,
			Concurrency: *concurrency,
			Banner:      *banner,
			Probe:       *probeFlag,
			ProbeTypes:  activeProbes,
		},
	}

	// Instanciar servicio
	svc := scan.NewService()

	// Ejecutar
	ctx := context.Background()

	reportResult, err := svc.Run(ctx, req)
	if err != nil {
		fmt.Printf("Scan failed: %v\n", err)
		os.Exit(1)
	}

	// Reportar
	report.PrintResults(reportResult.Results)

	fmt.Printf("Campaign completed in %v\n", reportResult.Metadata.Duration)
}
