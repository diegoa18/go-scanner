package orchestrator

import "time"

//define las reglas de negocio para el escaneo
type ScanPolicy struct {
	//comportamiento general
	Timeout     time.Duration
	Concurrency int

	ServiceDetection bool     //deteccion de servicios (pasiva o activa)
	ActiveProbing    bool     //probing activo (envio de payloads)
	AllowedProbes    []string //lista blanca de tipos de probes permitidos

	//AGREGAR MAS
}

//retorna una politica segura por defecto
func DefaultPolicy() ScanPolicy {
	return ScanPolicy{
		Timeout:          1 * time.Second,
		Concurrency:      100,
		ServiceDetection: true,  //detectar servicios de manera pasiva por defecto
		ActiveProbing:    false, //seguro por defecto
		AllowedProbes:    nil,
	}
}
