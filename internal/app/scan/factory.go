package scan

import (
	"fmt"
	"go-scanner/internal/model"
	"go-scanner/internal/orchestrator"
	"go-scanner/internal/scanner"
	"go-scanner/internal/scanner/tcp"
)

// factory para crear scanner
type ScannerFactory func(target string, ports []int, policy orchestrator.ScanPolicy, meta *model.HostMetadata) (scanner.Scanner, error)

// nueva instancia
func NewScanner(target string, ports []int, policy orchestrator.ScanPolicy, meta *model.HostMetadata) (scanner.Scanner, error) {
	switch policy.Type {
	case orchestrator.ScanTypeConnect:
		//TCP connect estandar
		return tcp.NewTCPConnectScanner(
			target,
			ports,
			policy.Timeout,
			policy.Concurrency,
			//ServiceDetection para activar el Banner Grabbing
			policy.ServiceDetection,
			meta,
		), nil

	case orchestrator.ScanTypeSYN:
		//PENDIENTE -> validar privilegios para SYN
		return nil, fmt.Errorf("SYN scan not implemented yet")

	default:
		return nil, fmt.Errorf("unsupported scan type: %s", policy.Type)
	}
}
