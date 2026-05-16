package orchestrator

import (
	"go-scanner/internal/discover/policy"
	"time"
)

// tipo de ecaneo
type ScanType string

const (
	ScanTypeConnect ScanType = "CONNECT"
	ScanTypeSYN     ScanType = "SYN"
	ScanTypeUDP     ScanType = "UDP"
)

// define las reglas de negocio para el escaneo
type ScanPolicy struct {
	Type ScanType
	//comportamiento general
	Timeout     time.Duration
	Concurrency int

	ServiceDetection bool     //deteccion de servicios (pasiva o activa)
	ActiveProbing    bool     //probing activo (envio de payloads)
	AllowedProbes    []string //lista blanca de tipos de probes permitidos

	// Politica de descubrimiento (fase previa)
	Discovery policy.Policy
}


