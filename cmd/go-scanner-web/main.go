package main

import (
	"go-scanner/cmd/go-scanner-web/app/server"
	"log"
)

func main() {
	port := 8080
	templateDir := "cmd/go-scanner-web/app/views/templates"

	// iniciar servidor
	if err := server.Start(port, templateDir); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
