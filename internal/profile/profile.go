package profile

import (
	"go-scanner/internal/orchestrator"
	"time"
)

// representa un preset de escaneo
type Profile struct {
	Name        string
	Description string
	Policy      orchestrator.ScanPolicy
}

// perfiles predefinidos
var (
	//PASSIVE: escaneo sigiloso, solo deteccion pasiva
	Passive = Profile{
		Name:        "passive",
		Description: "Passive scan: no active probing, only port detection and banner grabbing",
		Policy: orchestrator.ScanPolicy{
			Timeout:          2 * time.Second,
			Concurrency:      50,
			ServiceDetection: true,
			ActiveProbing:    false,
			AllowedProbes:    nil,
		},
	}

	// DEFAULT: escaneo balanceado, seguro por defecto
	Default = Profile{
		Name:        "default",
		Description: "Balanced scan: service detection enabled, no active probing by default",
		Policy: orchestrator.ScanPolicy{
			Timeout:          1 * time.Second,
			Concurrency:      100,
			ServiceDetection: true,
			ActiveProbing:    false,
			AllowedProbes:    nil,
		},
	}

	// AGGRESSIVE: escaneo rapido con probing activo
	Aggressive = Profile{
		Name:        "aggressive",
		Description: "Aggressive scan: faster, active probing enabled on HTTP/HTTPS",
		Policy: orchestrator.ScanPolicy{
			Timeout:          500 * time.Millisecond,
			Concurrency:      200,
			ServiceDetection: true,
			ActiveProbing:    true,
			AllowedProbes:    []string{"http", "https"},
		},
	}
)

// REGISTRO -> perfiles disponibles
var registry = map[string]Profile{
	"passive":    Passive,
	"default":    Default,
	"aggressive": Aggressive,
}

// retorna un perfil a partir de su nombre
func Get(name string) (Profile, bool) {
	p, ok := registry[name]
	return p, ok
}

// retorna nombres de perfiles disponibles
func Available() []string {
	return []string{"passive", "default", "aggressive"}
}
