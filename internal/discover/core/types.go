package core

import (
	"context"
	"go-scanner/internal/model"
	"time"
)

// modelo neutral reusable para resultados de descubrimiento
type HostResult struct {
	IP         string
	Alive      bool
	RTT        time.Duration
	Method     string                // icmp, tcp-connect
	Reason     string                // timeout, refused, echo-reply, etc...
	Error      error                 // internal error
	Timestamp  time.Time             // when it occurred
	Confidence model.ConfidenceLevel // calculated confidence
	Score      float64               // numeric score
}

// interfaz que deben implementar los discoverers
type Discoverer interface {
	Discover(ctx context.Context, target string) (HostResult, error)
}
