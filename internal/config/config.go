package config

import (
	"flag"
	"fmt"
	"go-scanner/internal/utils"
	"time"
)

// config almacena la configuración global del escaner (TEMPORAL)
type Config struct {
	Target      string        //host o ip objetivo
	PortRange   string        //rango de puertos (string debido a comas)
	Ports       []int         //slice de puertos parseados listos para usar
	Timeout     time.Duration //tiempo maximo de espera por conexión
	Concurrency int           //numero de gorutinas concurrentes
}

// globalconfig instancia global para acceso fácil (opcional, dependiendo del diseño final).
var GlobalConfigConfig *Config

// load procesa los argumentos de línea de comandos y retorna una configuración validada
func Load() (*Config, error) {
	target := flag.String("target", "", "target IP or Host")                                     //target es el host o ip objetivo (-target)
	portRange := flag.String("ports", "1-1024", "Ports to scan (e.g., '80', '1-100', '80,443')") //portRange es el rango de puertos a escanear (-ports)
	timeoutMs := flag.Int("timeout", 1000, "Timeout per connection in milliseconds")             //timeoutMs es el timeout por conexion en milisegundos (-timeout)
	concurrency := flag.Int("threads", 100, "Maximum number of concurrent connections")          //concurrency es el numero maximo de conexiones concurrentes (-threads)

	flag.Parse()

	if *target == "" { //validacion del target
		return nil, fmt.Errorf("target IP or Host is required")
	}

	if !utils.IsValidIP(*target) {
		//PENDIENTE -> implementar resolucion DNS
	}

	ports, err := utils.ParsePortRange(*portRange) //parseo de puertos
	if err != nil {
		return nil, fmt.Errorf("error parsing ports: %w", err)
	}

	cfg := &Config{ //contruccion de objeto config
		Target:      *target,
		PortRange:   *portRange,
		Ports:       ports,
		Timeout:     time.Duration(*timeoutMs) * time.Millisecond,
		Concurrency: *concurrency,
	}

	return cfg, nil
}
