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

// maneja el comando top "tcp" con suss sub -> connect, syn
func handleTCPCommand(args []string) {
	if len(args) < 1 {
		printTCPUsage()
		os.Exit(1)
	}

	subcommand := args[0]
	switch subcommand {
	case "connect":
		handleTCPGeneric(args[1:], "CONNECT")
	case "syn":
		handleTCPGeneric(args[1:], "SYN")
	default:
		fmt.Printf("Subcommand unknown for tcp: %s\n", subcommand)
		printTCPUsage()
		os.Exit(1)
	}
}

func printTCPUsage() {
	fmt.Println("Usage: go-scanner tcp <subcommand> [options] <target>")
	fmt.Println("Subcommands available:")
	fmt.Println("  connect    Perform a complete TCP Connect scan (User mode)")
	fmt.Println("  syn        Perform a Stealth TCP SYN scan (Root/CAP_NET_RAW required)")
}

// logica generica para escaneos TCP (connect, syn)
func handleTCPGeneric(args []string, scanType string) {
	dispName := strings.ToLower(scanType)
	cmd := flag.NewFlagSet(dispName, flag.ExitOnError)

	//flags comunes
	profileName := cmd.String("profile", "default", "Scan profile: passive, default, aggressive")
	portRange := cmd.String("p", "1-1024", "Ports to scan (e.g: '80', '1-1024', '80,443')")
	timeoutMs := cmd.Int("timeout", -1, "Timeout per connection in ms (default: from profile)")
	concurrency := cmd.Int("threads", -1, "Maximum number of concurrent connections (default: from profile)")

	//flags irrelevantes para SYN
	banner := cmd.Bool("banner", false, "Enable passive banner grabbing (Connect scan only)")
	probeFlag := cmd.Bool("probe", false, "Enable ACTIVE probing on detected services")
	probeTypes := cmd.String("probe-types", "http,https", "Comma-separated list of probe types to run (default: http,https)")
	allPorts := cmd.Bool("all", false, "Show all scanned ports (including CLOSED)")

	cmd.Parse(args)

	//validar argumentos
	if cmd.NArg() < 1 {
		fmt.Println("Error: target required (IP, CIDR or Range)")
		fmt.Printf("Usage: go-scanner tcp %s -p <ports> <target>\n", dispName)
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

	//configurar request con el ScanType explicito
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
			ScanType:    scanType, //inyeccion critica
		},
	}

	// Instanciar servicio
	svc := scan.NewService()

	// Ejecutar
	ctx := context.Background()

	reportResult, err := svc.Run(ctx, req)
	if err != nil {
		fmt.Printf("Scan failed: %v\n", err)

		if strings.Contains(err.Error(), "privileged") || strings.Contains(err.Error(), "scans are only supported") {
			if os.Geteuid() != 0 {
				fmt.Println("HINT: This scan type likely requires root privileges. Try with sudo.")
			}
		}
		os.Exit(1)
	}

	// Reportar
	report.PrintResults(reportResult.Results, *allPorts)

	fmt.Printf("Campaign completed in %v\n", reportResult.Metadata.Duration)
}
