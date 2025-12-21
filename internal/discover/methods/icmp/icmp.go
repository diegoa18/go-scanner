package icmp

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"go-scanner/internal/model"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// usa paquetes ICMP Echo Reply para detectar hosts
type Discoverer struct {
	Timeout time.Duration
}

// resultado del descubrimiento (definido localmente para evitar ciclo)
type HostResult struct {
	IP         string
	Alive      bool
	RTT        time.Duration
	Method     string
	Reason     string
	Error      error
	Timestamp  time.Time
	Confidence model.ConfidenceLevel
	Score      float64
}

// nueva instacia con timeout default
func NewDiscoverer(timeout time.Duration) *Discoverer {
	return &Discoverer{
		Timeout: timeout,
	}
}

// envia ping
func (d *Discoverer) Discover(ctx context.Context, target string) (HostResult, error) {
	result := HostResult{
		IP:        target,
		Alive:     false,
		Method:    "icmp",
		Timestamp: time.Now(),
	}

	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		result.Error = err
		result.Reason = "socket-error"
		return result, fmt.Errorf("icmp listen error: %w", err)
	}
	defer c.Close()

	//resolver IP
	dst, err := net.ResolveIPAddr("ip4", target)
	if err != nil {
		result.Error = err
		result.Reason = "dns-error"
		return result, err
	}

	// construir mensaje ICMP Echo
	m := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte("GO-SCANNER-HELLO"),
		},
	}
	b, err := m.Marshal(nil)
	if err != nil {
		result.Error = err
		return result, err
	}

	// enviar
	start := time.Now()
	if _, err := c.WriteTo(b, dst); err != nil {
		result.Error = err
		result.Reason = "send-error"
		return result, err
	}

	//leer respuesta con timeout/contexto
	reply := make([]byte, 1500)

	//configurar deadline
	deadline := time.Now().Add(d.Timeout)
	if err := c.SetReadDeadline(deadline); err != nil {
		result.Error = err
		return result, err
	}

	// loop de lectura
	for {
		select {
		case <-ctx.Done():
			result.Reason = "context-canceled"
			return result, ctx.Err()
		default:
		}

		n, peer, err := c.ReadFrom(reply)
		if err != nil {
			//timeout o error
			result.Reason = "timeout" //asumir timeout si no leyo
			return result, nil        //retornamos Alive=false, no fatal error
		}

		//validar source
		if peer.String() != dst.String() {
			continue
		}

		// parsear mensaje
		rm, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), reply[:n])
		if err != nil {
			continue
		}

		switch rm.Type {
		case ipv4.ICMPTypeEchoReply:
			result.Alive = true
			result.RTT = time.Since(start)
			result.Reason = "echo-reply"
			return result, nil
		}
	}
}
