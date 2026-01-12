package scan

// define el estado de una operacion de escaneo
type ScanStatus string

const (
	StatusPending   ScanStatus = "pending"   //creado pero no iniciado
	StatusRunning   ScanStatus = "running"   //en ejecucion
	StatusCompleted ScanStatus = "completed" //finalizado exitosamente
	StatusPartial   ScanStatus = "partial"   //finalizado con errores no fatales
	StatusFailed    ScanStatus = "failed"    //finalizado con errores fatales
)
