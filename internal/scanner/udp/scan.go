package udp

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"

	"go-scanner/internal/model"
	"go-scanner/internal/scanner"
)

// ensure UDPScanner implements scanner.Scanner
var _ scanner.Scanner = (*UDPScanner)(nil)

type UDPScanner struct {
	Target      string
	Ports       []int
	Timeout     time.Duration
	Concurrency int
	Metadata    *model.HostMetadata
}

func NewUDPScanner(target string, ports []int, timeout time.Duration, concurrency int, meta *model.HostMetadata) *UDPScanner {
	return &UDPScanner{
		Target:      target,
		Ports:       ports,
		Timeout:     timeout,
		Concurrency: concurrency,
		Metadata:    meta,
	}
}

func (s *UDPScanner) Scan(results chan<- scanner.ScanResult) {
	defer close(results)

	dstIP := net.ParseIP(s.Target).To4()
	if dstIP == nil {
		results <- scanner.ScanResult{
			Host:  s.Target,
			Port:  0,
			Error: fmt.Errorf("invalid IPv4 target"),
		}
		return
	}

	hasPrivileges := s.checkPrivileges()

	var scanWg sync.WaitGroup
	resultsMap := make(map[int]scanner.ScanResult)
	mu := sync.Mutex{}

	sem := make(chan struct{}, s.Concurrency)

	for _, port := range s.Ports {
		scanWg.Add(1)
		sem <- struct{}{}

		go func(p int) {
			defer scanWg.Done()
			defer func() { <-sem }()

			state := s.scanPort(p)

			mu.Lock()
			resultsMap[p] = scanner.ScanResult{
				Host:     s.Target,
				Port:     p,
				State:    state,
				Metadata: s.Metadata,
			}
			mu.Unlock()
		}(port)
	}

	scanWg.Wait()

	if hasPrivileges {
		var icmpWg sync.WaitGroup
		icmpWg.Add(1)
		go func() {
			defer icmpWg.Done()
			s.updateClosedPortsWithICMP(dstIP, resultsMap)
		}()
		icmpWg.Wait()
	}

	for _, port := range s.Ports {
		if res, ok := resultsMap[port]; ok {
			results <- res
		} else {
			results <- scanner.ScanResult{
				Host:     s.Target,
				Port:     port,
				State:    scanner.PortStateFiltered,
				Metadata: s.Metadata,
			}
		}
	}
}

func (s *UDPScanner) scanPort(port int) scanner.PortState {
	address := net.JoinHostPort(s.Target, fmt.Sprintf("%d", port))

	conn, err := net.DialTimeout("udp", address, s.Timeout)

	if err != nil {
		return scanner.PortStateFiltered
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(s.Timeout))

	n, err := conn.Read(buf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return scanner.PortStateOpen
		}
		return scanner.PortStateFiltered
	}

	if n > 0 {
		return scanner.PortStateOpen
	}

	return scanner.PortStateOpen
}

func (s *UDPScanner) checkPrivileges() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	if os.Geteuid() != 0 {
		return false
	}
	return true
}

func (s *UDPScanner) updateClosedPortsWithICMP(dstIP net.IP, resultsMap map[int]scanner.ScanResult) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
	if err != nil {
		return
	}
	defer syscall.Close(fd)

	buffer := make([]byte, 65535)
	deadline := time.Now().Add(500 * time.Millisecond)

	for {
		if time.Now().After(deadline) {
			break
		}

		setimeout := &syscall.Timeval{Sec: 0, Usec: 500000}
		syscall.SetsockoptTimeval(fd, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, setimeout)

		n, _, err := syscall.Recvfrom(fd, buffer, 0)
		if err != nil {
			continue
		}

		if n < 28 {
			continue
		}

		ipHeaderLen := (buffer[0] & 0x0F) * 4
		if int(ipHeaderLen) > n || ipHeaderLen < 20 {
			continue
		}

		capturedSrcIP := net.IP(buffer[12:16])
		if !capturedSrcIP.Equal(dstIP) {
			continue
		}

		icmpType := buffer[ipHeaderLen]
		if icmpType != 3 {
			continue
		}

		icmpData := buffer[ipHeaderLen+8:]
		if len(icmpData) < 8 {
			continue
		}

		originalIPHeader := icmpData[4:]
		if len(originalIPHeader) < 12 {
			continue
		}

		dstPort := uint16(originalIPHeader[10])<<8 | uint16(originalIPHeader[11])

		port := int(dstPort)
		if res, exists := resultsMap[port]; exists {
			if res.State == scanner.PortStateOpen || res.State == scanner.PortStateFiltered {
				resultsMap[port] = scanner.ScanResult{
					Host:     s.Target,
					Port:     port,
					State:    scanner.PortStateClosed,
					Metadata: s.Metadata,
				}
			}
		}
	}
}


