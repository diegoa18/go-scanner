package scan

import (
	"fmt"
	"time"
)

// error general durante el ciclo de un escaneo
type ScanError struct {
	Phase     string    //fase
	Err       error     //error original
	Target    string    //target asociado al error
	Timestamp time.Time //momento de este error
}

func (e *ScanError) Error() string {
	if e.Target != "" {
		return fmt.Sprintf("[%s] error en target %s: %v", e.Phase, e.Target, e.Err)
	}
	return fmt.Sprintf("[%s] error: %v", e.Phase, e.Err)
}

// nueva instancia de un error
func NewScanError(phase string, err error, target string) ScanError {
	return ScanError{
		Phase:     phase,
		Err:       err,
		Target:    target,
		Timestamp: time.Now(),
	}
}
