package tcp

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"go-scanner/internal/model"
	"go-scanner/internal/scanner"
	"math/rand"
	"net"
	"syscall"
	"time"
)

//TCP SYN SCANNER!

// estructura del SYN
type TCPSynScanner struct {
	Target      string
	Ports       []int
	Timeout     time.Duration
	Concurrency int
	Metadata    *model.HostMetadata
}

// representacion de los 20 bytes del header TCP
type TCPHeader struct {
	Source      uint16
	Destination uint16
	SeqNum      uint32
	AckNum      uint32
	DataOffset  uint8 // 4b altos -> Data Offset, 4b bajos -> Reserved
	Flags       uint8 // CWR, ECE, URG, ACK, PSH, RST, SYN, FIN
	Window      uint16
	Checksum    uint16
	Urgent      uint16
}

// (NO SE ENVIA POR LA RED, ES PARA EL CHECKSUM)
type PseudoHeader struct {
	SourceIP      [4]byte
	DestinationIP [4]byte
	Zero          uint8
	Protocol      uint8
	TCPLength     uint16
}

// Nueva instancia de TCPSynScanner
func NewTCPSynScanner(target string, ports []int, timeout time.Duration, concurrency int, meta *model.HostMetadata) *TCPSynScanner {
	return &TCPSynScanner{
		Target:      target,
		Ports:       ports,
		Timeout:     timeout,
		Concurrency: concurrency,
		Metadata:    meta,
	}
}

// corazon del SYN scanner
func (s *TCPSynScanner) Scan(results chan<- scanner.ScanResult) {
	defer close(results)

	// ressolver IP
	dstIP := net.ParseIP(s.Target).To4()
	if dstIP == nil {
		s.reportFatalError(results, s.Ports[0], fmt.Errorf("invalid IPv4 target"))
		return
	}

	// IP local -> para el checksum
	srcIP, err := getLocalIP(dstIP)
	if err != nil {
		s.reportFatalError(results, s.Ports[0], fmt.Errorf("failed to get local IP: %v", err))
		return
	}

	//creacion del socket RAW -> permisos root
	// AF_INET -> IPv4, SOCK_RAW -> Acceso crudo, IPPROTO_TCP -> Protocolo TCP
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_TCP)
	if err != nil {
		s.reportFatalError(results, s.Ports[0], fmt.Errorf("raw socket creation failed (are you root?): %v", err))
		return
	}
	defer syscall.Close(fd)

	//traking de puertos enviados
	sentPorts := make(map[uint16]time.Time)
	for _, p := range s.Ports {
		sentPorts[uint16(p)] = time.Now()
	}

	//contexto con timeout (evita que el listener sea infinito)
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout+(2*time.Second)) // Timeout + margen
	defer cancel()

	found := make(chan scanner.ScanResult, len(s.Ports)) //canal para resultados encontrados

	//listener (gorutine)
	go s.listen(ctx, fd, dstIP, sentPorts, found)

	//sender (consstruccion de TCP SYN)
	s.sendPackets(fd, dstIP, srcIP)

	//ressultados
	resultsMap := make(map[int]scanner.ScanResult)

	//essperar el contexto (el timeout)
	<-ctx.Done()
	close(found)

	for res := range found {
		resultsMap[res.Port] = res
	}

	for _, port := range s.Ports {
		if res, ok := resultsMap[port]; ok {
			results <- res

		} else { //sin resspuesta -> FILTERED
			results <- scanner.ScanResult{
				Host:     s.Target,
				Port:     port,
				State:    scanner.PortStateFiltered,
				Banner:   "",
				Metadata: s.Metadata,
			}
		}
	}
}

// envio de paquetes TCP SYN
func (s *TCPSynScanner) sendPackets(fd int, dstIP net.IP, srcIP net.IP) {
	// socket address estructura para syscall
	sa := &syscall.SockaddrInet4{Port: 0}
	copy(sa.Addr[:], dstIP)

	// simulacion de trafico real (puerto fuente aleatorio (o eso intento))
	srcPort := uint16(1024 + rand.Intn(60000))

	for _, port := range s.Ports {
		dstPort := uint16(port)

		// contruccion real de TCP SYN para el envio
		tcpH := TCPHeader{
			Source:      srcPort,
			Destination: dstPort,
			SeqNum:      rand.Uint32(),
			AckNum:      0,
			DataOffset:  5 << 4, // 20 bytes (5 words)
			Flags:       0x02,   // SYN flag set
			Window:      1024,
			Checksum:    0,
			Urgent:      0,
		}

		// calculo de checksum
		payload := tcpToBytes(&tcpH)
		tcpH.Checksum = calculateChecksum(payload, srcIP, dstIP)

		// re-serializacion con checksum correcto
		finalPacket := tcpToBytes(&tcpH)

		//envio (socket, segmento TCP, flags, socket address)
		err := syscall.Sendto(fd, finalPacket, 0, sa)
		if err != nil {
			//PEDIENTE
			continue
		}

		// pequeño control de flujo
		if s.Concurrency > 0 && s.Concurrency < 1000 {
			time.Sleep(time.Millisecond * 1)
		}
	}
}

// escucha respuestas de socket raw
func (s *TCPSynScanner) listen(ctx context.Context, fd int, targetIP net.IP, sentPorts map[uint16]time.Time, found chan<- scanner.ScanResult) {
	buffer := make([]byte, 4096) // buffer de lectura

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// leer ssocket
			n, _, err := syscall.Recvfrom(fd, buffer, 0)
			if err != nil {
				continue
			}

			// paquete capturado
			if n < 20+20 { //minimo IPv4 + TCP
				continue
			}

			// extraer IP Header length y asi encontrar TCP
			ipHeaderLen := (buffer[0] & 0x0F) * 4
			if int(ipHeaderLen) > n {
				continue
			}

			// ver IP Origen == Target
			capturedSrcIP := net.IP(buffer[12:16])
			if !capturedSrcIP.Equal(targetIP) {
				continue
			}

			// extraer Header TCP
			tcpBytes := buffer[ipHeaderLen:]
			if len(tcpBytes) < 20 {
				continue
			}

			var tcpH TCPHeader
			reader := bytes.NewReader(tcpBytes)
			if err := binary.Read(reader, binary.BigEndian, &tcpH); err != nil {
				continue
			}

			// TCP Source Port == scanned port
			if _, expected := sentPorts[tcpH.Source]; !expected {
				continue
			}

			// analizar flags -> SYN=0x02, ACK=0x10, RST=0x04
			var state scanner.PortState

			if (tcpH.Flags & 0x12) == 0x12 { // SYN + ACK
				state = scanner.PortStateOpen
			} else if (tcpH.Flags & 0x04) != 0 { // RST
				state = scanner.PortStateClosed
			} else {
				continue // otro paquetes
			}

			found <- scanner.ScanResult{
				Host:     s.Target,
				Port:     int(tcpH.Source),
				State:    state,
				Banner:   "", // SYN no captura banners
				Metadata: s.Metadata,
			}
		}
	}
}

// HELPERS
func sAddr(ip net.IP) [4]byte {
	var ret [4]byte
	copy(ret[:], ip.To4())
	return ret
}

func tcpToBytes(h *TCPHeader) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, h)
	return buf.Bytes()
}

func calculateChecksum(data []byte, src, dst net.IP) uint16 {
	pseudo := new(bytes.Buffer)
	binary.Write(pseudo, binary.BigEndian, sAddr(src))
	binary.Write(pseudo, binary.BigEndian, sAddr(dst))
	binary.Write(pseudo, binary.BigEndian, uint8(0))
	binary.Write(pseudo, binary.BigEndian, uint8(syscall.IPPROTO_TCP))
	binary.Write(pseudo, binary.BigEndian, uint16(len(data)))

	totalLen := pseudo.Len() + len(data)
	if totalLen%2 != 0 {
		totalLen++
	}

	buf := make([]byte, totalLen)
	copy(buf, pseudo.Bytes())
	copy(buf[pseudo.Len():], data)

	var sum uint32
	for i := 0; i < len(buf)-1; i += 2 {
		sum += uint32(binary.BigEndian.Uint16(buf[i : i+2]))
	}

	if len(buf)%2 != 0 {
		sum += uint32(uint16(buf[len(buf)-1]) << 8)
	}

	for sum>>16 != 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}

	return ^uint16(sum)
}

func getLocalIP(dst net.IP) (net.IP, error) {
	// udp dummy para ver que interfaz elige el kernel
	conn, err := net.Dial("udp", dst.String()+":80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}

// reportar error fatal
func (s *TCPSynScanner) reportFatalError(results chan<- scanner.ScanResult, port int, err error) {
	results <- scanner.ScanResult{
		Host:     s.Target,
		Port:     port,
		Error:    err,
		State:    scanner.PortStateClosed,
		Metadata: s.Metadata,
	}
}
