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

func handleUDPCommand(args []string) {
	if len(args) < 1 {
		printUDPUsage()
		os.Exit(1)
	}

	handleUDPScan(args)
}

func printUDPUsage() {
	fmt.Println("Usage: go-scanner udp [options] <target>")
	fmt.Println("Options:")
	fmt.Println("  -p <ports>       Ports to scan (e.g: '53,67,123,161' or '1-1000')")
	fmt.Println("  --profile        Scan profile: passive, default, aggressive")
	fmt.Println("  --timeout        Timeout per packet in ms")
	fmt.Println("  --threads        Maximum concurrent packets")
	fmt.Println("  --all            Show all scanned ports")
	fmt.Println("\nExample:")
	fmt.Println("  go-scanner udp -p 53,67,123 192.168.1.1")
}

func handleUDPScan(args []string) {
	cmd := flag.NewFlagSet("udp", flag.ExitOnError)

	profileName := cmd.String("profile", "default", "Scan profile: passive, default, aggressive")
	portRange := cmd.String("p", "53,67,123,161,500,4500", "Ports to scan")
	timeoutMs := cmd.Int("timeout", -1, "Timeout per packet in ms")
	concurrency := cmd.Int("threads", -1, "Maximum concurrent packets")
	allPorts := cmd.Bool("all", false, "Show all scanned ports")

	cmd.Parse(args)

	if cmd.NArg() < 1 {
		fmt.Println("Error: target required")
		fmt.Printf("Usage: go-scanner udp -p <ports> <target>\n")
		os.Exit(1)
	}

	rawTarget := cmd.Arg(0)

	targets, err := utils.ParseTarget(rawTarget)
	if err != nil {
		fmt.Printf("Error validating target: %v\n", err)
		os.Exit(1)
	}

	if len(targets) == 0 {
		fmt.Println("Error: No valid targets found.")
		os.Exit(1)
	}

	req := scan.ScanRequest{
		Targets:     targets,
		Ports:       *portRange,
		ProfileName: *profileName,
		Options: scan.ScanOptions{
			TimeoutMs:   *timeoutMs,
			Concurrency: *concurrency,
			ScanType:    "UDP",
		},
	}

	svc := scan.NewService()
	ctx := context.Background()

	reportResult, err := svc.Run(ctx, req)
	if err != nil {
		fmt.Printf("Scan failed: %v\n", err)

		if strings.Contains(err.Error(), "privileged") || strings.Contains(err.Error(), "root") {
			if os.Geteuid() != 0 {
				fmt.Println("HINT: UDP scan requires root privileges. Try with sudo.")
			}
		}
		os.Exit(1)
	}

	report.PrintResults(reportResult.Results, *allPorts)

	fmt.Printf("Scan completed in %v\n", reportResult.Metadata.Duration)
}
