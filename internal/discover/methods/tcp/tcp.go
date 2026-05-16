package tcp

import (
	"context"
	"fmt"
	"go-scanner/internal/model"
	"net"
	"strings"
	"time"
)

type ConnectDiscoverer struct {
	Ports   []int
	Timeout time.Duration
}

func NewConnectDiscoverer(ports []int, timeout time.Duration) *ConnectDiscoverer {
	if len(ports) == 0 {
		ports = []int{80, 443}
	}
	return &ConnectDiscoverer{
		Ports:   ports,
		Timeout: timeout,
	}
}

func (d *ConnectDiscoverer) Discover(ctx context.Context, target string) (model.HostResult, error) {
	result := model.HostResult{
		IP:        target,
		Alive:     false,
		Method:    "tcp-connect",
		Timestamp: time.Now(),
	}

	for _, port := range d.Ports {
		address := fmt.Sprintf("%s:%d", target, port)

		dialer := net.Dialer{
			Timeout: d.Timeout,
		}

		start := time.Now()
		conn, err := dialer.DialContext(ctx, "tcp", address)

		if err == nil {
			conn.Close()
			result.Alive = true
			result.RTT = time.Since(start)
			result.Reason = fmt.Sprintf("syn-ack_port-%d", port)
			return result, nil
		}

		if isConnectionRefused(err) {
			result.Alive = true
			result.RTT = time.Since(start)
			result.Reason = fmt.Sprintf("rst_port-%d", port)
			return result, nil
		}
	}

	result.Reason = "timeout"
	return result, nil
}

func isConnectionRefused(err error) bool {
	s := err.Error()
	return strings.Contains(s, "refused") || strings.Contains(s, "reset")
}
