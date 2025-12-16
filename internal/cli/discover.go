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

	// instancia el discoverer
	timeout := time.Duration(*timeoutMs) * time.Millisecond
	d := discover.NewICMPDiscoverer(timeout)

	fmt.Printf("Starting ICMP Discovery on %s ... (Timeout: %v)\n", target, timeout)

	// ejecuta el descubrimiento
	ctx := context.Background()
	result, err := d.Discover(ctx, target)

	if err != nil {
		fmt.Printf("Error during discovery: %v\n", err)
		fmt.Println("NOTE: ICMP usually requires root/admin privileges.")
		os.Exit(1)
	}

	// muestrar el resultado
	if result.Alive {
		fmt.Printf("[ALIVE] %s (RTT: %v)\n", result.IP, result.RTT)
	} else {
		fmt.Printf("[DEAD] %s (or blocked)\n", result.IP)
	}
}

// helper stub para Windows (Geteuid no existe en windows, usar os.Getuid si existe o constante -1)
// En Go, os.Geteuid() retorna -1 en windows, asi que funciona "bien" para no crashear, pero no detecta admin real.
// Se dejara asi por simplicidad del ejemplo.
