package dto

// define los datos enviados por el formulario web
type ScanRequest struct {
	Target string
}

// define la estructura de datos para renderizar la pagina
type ScanPageData struct {
	Result    interface{}
	Error     string
	IsLoading bool
}
