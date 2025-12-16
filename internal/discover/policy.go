package discover

import "time"

//define la conf para la fase de descubrimiento
type Policy struct {
	Enabled     bool
	Methods     []string      // ICMP, TCP-Connect
	Timeout     time.Duration // timeout
	MaxHosts    int           // limite maximo de targets
	Concurrency int           // tamaño de pool de workers
	Delay       time.Duration // delay opcional entre targets
}
