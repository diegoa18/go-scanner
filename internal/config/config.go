package config

import (
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
}
