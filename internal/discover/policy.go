package discover

import "time"

//define la conf para la fase de descubrimiento
type Policy struct {
	Enabled bool
	Methods []string      // ICMP, TCP-Connect
	Timeout time.Duration //timeout
}
