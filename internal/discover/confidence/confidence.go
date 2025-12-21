package confidence

import (
	"fmt"
	"go-scanner/internal/model"
	"strings"
)

// señal individual de descubrimiento
type ResultSignal struct {
	Method string
	Alive  bool
	Reason string
	Error  error
}

// agrupa el resultado del motor de confianza
type CalculationResult struct {
	Level  model.ConfidenceLevel
	Score  float64
	Reason string
}

// define los pesos para cada tipo de señal
var (
	ScoreICMPReply  = 10.0
	ScoreTCPOpen    = 10.0
	ScoreTCPClosed  = 5.0 // RST indica host vivo, pero puerto cerrado
	ScoreTimeout    = 0.0
	ThumbnailHigh   = 15.0 // Un metodo fuerte + algo mas, o dos fuertes
	ThumbnailMedium = 5.0  // Al menos un metodo debil (RST)
)

// determina el nivel de confianza basado en una lista de señales
func Calculate(signals []ResultSignal) CalculationResult {
	if len(signals) == 0 {
		return CalculationResult{
			Level:  model.ConfidenceUnknown,
			Score:  0,
			Reason: "no-signals",
		}
	}

	var score float64
	var reasons []string
	aliveCount := 0

	for _, s := range signals {
		if !s.Alive {
			//PENDIENTE: revisar si hubo un error explicito o solo timeout
			continue
		}

		aliveCount++
		signalScore := 0.0

		// logica base: determinamos puntaje por metodo y razon
		switch s.Method {
		case "icmp":
			if s.Reason == "echo-reply" {
				signalScore = ScoreICMPReply
			}
		case "tcp-connect":
			if s.Reason == "syn-ack" || s.Reason == "open" {
				signalScore = ScoreTCPOpen
			} else if s.Reason == "rst" || s.Reason == "refused" {
				// RST confirma que hay IP stack respondiendo
				signalScore = ScoreTCPClosed
			}
		}

		// acumulacion
		score += signalScore
		reasons = append(reasons, fmt.Sprintf("%s(%s)=%.0f", s.Method, s.Reason, signalScore))
	}

	// determinacion de nivel
	level := model.ConfidenceLow

	if aliveCount == 0 {
		level = model.ConfidenceUnknown
		reasons = []string{"all-failed"}
	} else {
		if score >= ThumbnailHigh {
			level = model.ConfidenceHigh
		} else if score >= ThumbnailMedium {
			level = model.ConfidenceMedium
		} else {
			level = model.ConfidenceLow
		}
	}

	return CalculationResult{
		Level:  level,
		Score:  score,
		Reason: strings.Join(reasons, ", "),
	}
}
