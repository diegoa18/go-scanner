package discover

import (
	"context"
	"time"
)

// modelo neutral reusable para resultados de descubrimiento
type HostResult struct {
	IP        string
	Alive     bool
	RTT       time.Duration
	Method    string    // icmp, tcp-connect
	Reason    string    // timeout, refused, echo-reply, etc...
	Error     error     // internal error
	Timestamp time.Time // when it occurred
}

// interfaz que deben implementar los discoverers
type Discoverer interface {
	Discover(ctx context.Context, target string) (HostResult, error)
}
