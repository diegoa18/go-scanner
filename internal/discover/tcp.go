package discover

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

// intenta conectar a una lista de puertos comunes
// si alguno responde (Open) o rechaza activamente (Closed/RST), el host existe
type TCPConnectDiscoverer struct {
	Ports   []int
	Timeout time.Duration
}

// nueva instancia
func NewTCPConnectDiscoverer(ports []int, timeout time.Duration) *TCPConnectDiscoverer {
	if len(ports) == 0 {
		ports = []int{80, 443} //defaults
	}
	return &TCPConnectDiscoverer{
		Ports:   ports,
		Timeout: timeout,
	}
}

func (d *TCPConnectDiscoverer) Discover(ctx context.Context, target string) (HostResult, error) {
	result := HostResult{
		IP:        target,
		Alive:     false,
		Method:    "tcp-connect",
		Timestamp: time.Now(),
	}

	// iterar sobre puertos definidos
	for _, port := range d.Ports {
		address := fmt.Sprintf("%s:%d", target, port)

		dialer := net.Dialer{
			Timeout: d.Timeout,
		}

		start := time.Now()
		conn, err := dialer.DialContext(ctx, "tcp", address)

		if err == nil {
			// CONEXION EXITOSA -> Host Alive
			conn.Close()
			result.Alive = true
			result.RTT = time.Since(start)
			result.Reason = fmt.Sprintf("syn-ack_port-%d", port)
			return result, nil
		}

		// analizar error
		if isConnectionRefused(err) {
			result.Alive = true // host respondio RST
			result.RTT = time.Since(start)
			result.Reason = fmt.Sprintf("rst_port-%d", port)
			return result, nil
		}
	}

	result.Reason = "timeout"
	return result, nil
}

// detecta si el error es por rechazo activo
func isConnectionRefused(err error) bool {
	if opErr, ok := err.(*net.OpError); ok {
		//WIN: WSAECONNREFUSED (10061)
		//LINUX: syscall.ECONNREFUSED
		_ = opErr
	}
	//PENDIENTE -> mejorar deteccion robusta cross-platform
	//(actualmente si no conecto exitosamente, no contamos como vivo)
	s := err.Error()
	return len(s) > 0 && (strings.Contains(s, "refused") || strings.Contains(s, "reset"))
}
