package orchestrator

import (
	"go-scanner/internal/discover"
	"time"
)

// tipo de ecaneo
type ScanType string

const (
	ScanTypeConnect ScanType = "CONNECT"
	ScanTypeSYN     ScanType = "SYN"
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
	Discovery discover.Policy
}

// retorna una politica segura por defecto
// CONSIDERAR USAR profile.Default.Policy EN SU LUGAR
func DefaultPolicy() ScanPolicy {
	return ScanPolicy{
		Type:             ScanTypeConnect,
		Timeout:          1 * time.Second,
		Concurrency:      100,
		ServiceDetection: true,  //detectar servicios de manera pasiva por defecto
		ActiveProbing:    false, //seguro por defecto
		AllowedProbes:    nil,

		//SECCION DISCOVERY
		Discovery: discover.Policy{
			Enabled:     true,
			Methods:     []string{"icmp", "tcp-connect"},
			Timeout:     2 * time.Second,
			MaxHosts:    1000,
			Concurrency: 50,
			Delay:       0,
		},
	}
}
