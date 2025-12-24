package scan

// define los parametros de entrada para un escaneo
type ScanRequest struct {
	Targets     []string    //lista de targets
	Ports       string      //puertos a escanear
	ProfileName string      //perfil de configuracion
	Options     ScanOptions //opciones de tuning
}

// define las opciones de tuning fino
type ScanOptions struct {
	TimeoutMs   int  //timeout en ms
	Concurrency int  //nivel de concurrencia
	Banner      bool //habilita la captura de banners explícitamente
	Probe       bool //habilita el probing activo
	ProbeTypes  []string
}
