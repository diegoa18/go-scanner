package discover

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// usa paquetes ICMP Echo Reply para detectar hosts
type ICMPDiscoverer struct {
	Timeout time.Duration
}

// nueva instacia con timeout default
func NewICMPDiscoverer(timeout time.Duration) *ICMPDiscoverer {
	return &ICMPDiscoverer{
		Timeout: timeout,
	}
}

// envia ping
func (d *ICMPDiscoverer) Discover(ctx context.Context, target string) (HostResult, error) {
	result := HostResult{
		IP:     target,
		Alive:  false,
		Method: "icmp",
	}

	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0") //creacion del socket (requiere permisos)
	if err != nil {
		//probablemente falta de privilegios
		return result, fmt.Errorf("error listening for ICMP (root required?): %w", err)
	}
	defer c.Close()

	//resolver IP
	dst, err := net.ResolveIPAddr("ip4", target)
	if err != nil {
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
		return result, err
	}

	// enviar
	start := time.Now()
	if _, err := c.WriteTo(b, dst); err != nil {
		return result, err
	}

	// leer respuesta con timeout/contexto
	reply := make([]byte, 1500)

	// configurar deadline para ReadFrom
	deadline := time.Now().Add(d.Timeout)
	if err := c.SetReadDeadline(deadline); err != nil {
		return result, err
	}

	//loop de lectura
	for {
		//verificar el contexto antes de la lectura
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		n, peer, err := c.ReadFrom(reply)
		if err != nil {
			//timeout o error
			return result, nil //retornar ALIVE o DEAD segun el caso
		}

		//validar que sea de quien esperamos
		if peer.String() != dst.String() {
			continue
		}

		//parsear mensaje
		rm, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), reply[:n])
		if err != nil {
			continue
		}

		switch rm.Type {
		case ipv4.ICMPTypeEchoReply:
			//(POR AHORA) si recibimos EchoReply de la IP destino, se asume ALIVE
			result.Alive = true
			result.RTT = time.Since(start)
			return result, nil
		}
	}
}
