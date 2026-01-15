package scan

import (
	"fmt"
	"go-scanner/internal/model"
	"go-scanner/internal/orchestrator"
	"go-scanner/internal/scanner"
	"go-scanner/internal/scanner/tcp"
	"os"
	"runtime"
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
		//verificar privilegios antes de crear el scanner
		if err := checkPrivileges(); err != nil {
			return nil, fmt.Errorf("privileged scan required: %w", err)
		}

		return tcp.NewTCPSynScanner(
			target,
			ports,
			policy.Timeout,
			policy.Concurrency,
			meta,
		), nil

	default:
		return nil, fmt.Errorf("unsupported scan type: %s", policy.Type)
	}
}

// valida si el proceso tiene permisos
func checkPrivileges() error {
	//solo linux
	if runtime.GOOS != "linux" {
		return fmt.Errorf("raw socket scans are only supported on Linux")
	}

	//root
	if os.Geteuid() != 0 {
		return fmt.Errorf("root privileges required for SYN scan")
	}

	return nil
}
