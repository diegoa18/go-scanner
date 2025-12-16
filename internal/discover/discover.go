package discover

import (
	"context"
	"time"
)

// modelo neutral reusable para resultados de descubrimiento
type HostResult struct {
	IP     string
	Alive  bool
	RTT    time.Duration
	Method string //ahora es icmp, pero mas adelante se puede emplear tcp-ack, arp, etc
}

// interfaz que deben emplear los discoverers
type Discoverer interface {
	//descoverer hacia un target IP y retorna el resultado
	Discover(ctx context.Context, target string) (HostResult, error)
}
