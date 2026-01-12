package scan

import (
	"go-scanner/internal/scanner"
	"time"
)

// reporte del escaneo
type ScanReport struct {
	JobID    string               //cada operacion tiene un id unico
	Status   ScanStatus           //estado de la operacion
	Results  []scanner.ScanResult //resultados del escaneo
	Errors   []ScanError          //errores ocurridos durante la ejecucion
	Metadata ExecutionMetadata    //metadata del escaneo
}

// contexto
type ExecutionMetadata struct {
	Duration    time.Duration //duracion de la operacion
	TargetCount int           //cantidad de targets
	ProfileUsed string        //perfil utilizado
}
