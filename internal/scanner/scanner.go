package scanner

import "fmt"

//es el resultado del escaneo de un unico puerto
type ScanResult struct {
	Port    int
	IsOpen  bool
	Service string //nombre del servicio
	Banner  string //banner capturado
	Error   error
}

//retorna representacion legible del resultado
func (r ScanResult) String() string {
	status := "CLOSED"
	if r.IsOpen {
		status = "OPEN"
	}
	return fmt.Sprintf("Port %d: %s", r.Port, status)
}

//define el contrato para cualquier tipo de escaner
type Scanner interface {
	//ejecuta el escaneo sobre el target configurado y envia resultados al canal
	Scan(results chan<- ScanResult)
}
