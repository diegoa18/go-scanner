package probe

import "time"

// definer el comportamiento de un prober
type Prober interface {
	//ejecutar la prueba activa sobre una direccion y puerto
	Probe(target string, port int, timeout time.Duration) (string, error)
}
