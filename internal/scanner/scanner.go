package scanner

import (
	"fmt"
	"go-scanner/internal/model"
)

// ESTADO DEL PORT
type PortState string

const (
	PortStateOpen     PortState = "OPEN"
	PortStateClosed   PortState = "CLOSED"
	PortStateFiltered PortState = "FILTERED"
)

// es el resultado del escaneo de un unico puerto
type ScanResult struct {
	Host     string //IP o hostname
	Port     int
	State    PortState // Estado explicito del puerto
	Service  string    //nombre del servicio
	Banner   string    //banner capturado
	Error    error
	Metadata *model.HostMetadata //contexto del host discovery
}

// IsOpen helper
func (r ScanResult) IsOpen() bool {
	return r.State == PortStateOpen
}

// representacion bonita del resultado
func (r ScanResult) String() string {
	return fmt.Sprintf("[%s] Port %d: %s", r.Host, r.Port, r.State)
}

// define el contrato para cualquier tipo de escaner
type Scanner interface {
	//ejecuta el escaneo sobre el target configurado y envia resultados al canal
	Scan(results chan<- ScanResult)
}
