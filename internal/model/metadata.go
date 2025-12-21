package model

import "time"

// encapsula el contexto del descubrimiento sobre un target obtenido
type HostMetadata struct {
	ID              string          // IP
	DiscoveryMethod string          //metodo de descubrimiento (como icmp o tcp-connect)
	DiscoveryRTT    time.Duration   //tiempo de respuesta
	DiscoveryReason string          //razon de vida (syn-ack, echo-reply)
	DiscoveryTime   time.Time       //momento del descubrimiento
	Confidence      ConfidenceLevel //high, medium, low
}

// nivel de confianza del descubrimiento
type ConfidenceLevel string

const (
	ConfidenceHigh    ConfidenceLevel = "high"
	ConfidenceMedium  ConfidenceLevel = "medium"
	ConfidenceLow     ConfidenceLevel = "low"
	ConfidenceUnknown ConfidenceLevel = "unknown"
)
