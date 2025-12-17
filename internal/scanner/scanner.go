package scanner

import (
	"fmt"
	"go-scanner/internal/model"
)

// es el resultado del escaneo de un unico puerto
type ScanResult struct {
	Host     string //IP o hostname
	Port     int
	IsOpen   bool
	Service  string //nombre del servicio
	Banner   string //banner capturado
	Error    error
	Metadata *model.HostMetadata //contexto del host discovery
}

// retorna representacion legible del resultado
func (r ScanResult) String() string {
	status := "CLOSED"
	if r.IsOpen {
		status = "OPEN"
	}
	return fmt.Sprintf("[%s] Port %d: %s", r.Host, r.Port, status)
}

// define el contrato para cualquier tipo de escaner
type Scanner interface {
	//ejecuta el escaneo sobre el target configurado y envia resultados al canal
	Scan(results chan<- ScanResult)
}
