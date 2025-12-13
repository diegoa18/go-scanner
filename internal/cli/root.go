package cli

import (
	"fmt"
	"os"
)

// Execute es el punto de entrada para la logica del CLI
func Execute() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	//el primer argumento es el comando principal (como lo son tcp o udp)
	command := os.Args[1]

	//argumentos restantes para el subcomando
	subArgs := os.Args[2:]

	switch command {
	case "tcp":
		handleTCPCommand(subArgs) // Llama a la funcion en tcp.go (mismo paquete)
	/*
		AQUI AGREGAR MAS CASOS, TAMBIEN EVALUAR SI ES FACTIBLE USAR SWITCH CASE (NO LO CREOOOOO)
	*/
	default:
		fmt.Printf("unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

// mensajes de ayuda (MEJORAR)
func printUsage() {
	fmt.Println("Usage: go-scanner <command> <subcommand> [options] <target>")
	fmt.Println("Commands available:")
	fmt.Println("  tcp    TCP scan tools")
	fmt.Println("\nExample:")
	fmt.Println("  go-scanner tcp connect -p 80,443 192.168.1.1")
}
