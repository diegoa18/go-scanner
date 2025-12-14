package config

import (
	"fmt"
	"time"
)

// la configuracion de un unico escaneo
type Config struct {
	Target       string        // host o IP objetivo
	PortRange    string        // rango de puertos original (string)
	Ports        []int         // slice de puertos parseados listos para usar
	Timeout      time.Duration // tiempo maximo de espera por conexion
	Concurrency  int           // numero de gorutinas concurrentes
	EnableBanner bool          // habilitar banner grabbing pasivo
	EnableProbe  bool          // habilitar probing activo
	ProbeTypes   []string      // tipos de probes activos permitidos (default: todos o http)
}

// valida coherencia
func (c *Config) Validate() error {
	if c.Target == "" {
		return fmt.Errorf("target is required")
	}

	if len(c.Ports) == 0 {
		return fmt.Errorf("no ports to scan")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if c.Concurrency <= 0 {
		return fmt.Errorf("concurrency must be positive")
	}

	return nil
}
