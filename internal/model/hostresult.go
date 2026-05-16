package model

import "time"

type HostResult struct {
	IP         string
	Alive      bool
	RTT        time.Duration
	Method     string
	Reason     string
	Error      error
	Timestamp  time.Time
	Confidence ConfidenceLevel
	Score      float64
}
