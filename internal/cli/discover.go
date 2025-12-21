package cli

import (
	"context"
	"flag"
	"fmt"
	"go-scanner/internal/discover"
	"os"
	"time"
)

// maneja el comando "discover"
func handleDiscoverCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: go-scanner discover <subcommand> [options] <target>")
		fmt.Println("Subcommands available:")
		fmt.Println("  icmp    Perform ICMP Echo (ping) host discovery")
		os.Exit(1)
	}

	subcommand := args[0]
	switch subcommand {
	case "icmp":
		handleICMPDiscover(args[1:])
	default:
		fmt.Printf("Subcommand unknown for discover: %s\n", subcommand)
		os.Exit(1)
	}
}

// ejecuta la logica para descubrimiento ICMP
func handleICMPDiscover(args []string) {
	cmd := flag.NewFlagSet("icmp", flag.ExitOnError)
	timeoutMs := cmd.Int("timeout", 2000, "Timeout in ms")

	cmd.Parse(args)

	if cmd.NArg() < 1 {
		fmt.Println("Error: target required (IP)")
		fmt.Println("Usage: go-scanner discover icmp [-timeout ms] <target>")
		os.Exit(1)
	}
	target := cmd.Arg(0)

	// configurar policy para ICMP discovery
	timeout := time.Duration(*timeoutMs) * time.Millisecond
	policy := discover.Policy{
		Enabled:     true,
		Methods:     []string{"icmp"},
		Timeout:     timeout,
		MaxHosts:    0,
		Concurrency: 1,
		Delay:       0,
	}

	fmt.Printf("Starting ICMP Discovery on %s ... (Timeout: %v)\n", target, timeout)
	fmt.Println("NOTE: ICMP requires root/admin privileges. Run with 'sudo' if you see permission errors.")

	// ejecuta el descubrimiento
	ctx := context.Background()
	results, err := discover.Run(ctx, []string{target}, policy)

	if err != nil {
		fmt.Printf("Error during discovery: %v\n", err)
		fmt.Println("\n⚠️  ICMP Discovery requires elevated privileges!")
		fmt.Println("Try running with: sudo go run cmd/go-scanner/main.go discover icmp <target>")
		os.Exit(1)
	}

	// muestrar el resultado
	if len(results) > 0 && results[0].Alive {
		fmt.Printf("✓ [ALIVE] %s (RTT: %v, Confidence: %s)\n", results[0].IP, results[0].RTT, results[0].Confidence)
	} else {
		fmt.Printf("✗ [DEAD] %s\n", target)
		if len(results) > 0 {
			fmt.Printf("   Reason: %s\n", results[0].Reason)
			if results[0].Error != nil {
				fmt.Printf("   Error: %v\n", results[0].Error)
			}
		}
		fmt.Println("\n💡 Possible reasons:")
		fmt.Println("   1. Host is actually down")
		fmt.Println("   2. ICMP is blocked by firewall")
		fmt.Println("   3. You need root privileges (try with 'sudo')")
	}
}
