package scanner

import (
	"fmt"
	"net"  //API de red
	"sync" //sincronizacion
	"time"
)

// encapsula todo el estado necesario para realizar un escaneo TCP
type TCPConnectScanner struct {
	Target      string        //host o ip objetivo
	Ports       []int         //puertos a escanear
	Timeout     time.Duration //timeout por conexion
	Concurrency int           //numero maximo de conexiones concurrentes
}

// nueva instacia de TCPConnectScanner
func NewTCPConnectScanner(target string, ports []int, timeout time.Duration, concurrency int) *TCPConnectScanner {
	return &TCPConnectScanner{
		Target:      target,
		Ports:       ports,
		Timeout:     timeout,
		Concurrency: concurrency,
	}
}

// debe iterar sobre los puertos y lanzar gorutinas limitadas
func (s *TCPConnectScanner) Scan(results chan<- ScanResult) {
	defer close(results)

	var wg sync.WaitGroup
	// semaforo para limitar concurrencia y no saturar FDs o la red.
	sem := make(chan struct{}, s.Concurrency)

	for _, port := range s.Ports { //recorrer cada puerto
		wg.Add(1)
		sem <- struct{}{} //intenta adquirir un slot del semaforo
		//si el semaforo esta lleno, el loop se bloquea y se evita crear mas gorutinas

		go func(p int) {
			defer wg.Done()
			defer func() { <-sem }() //libera el slot del semaforo

			isOpen := s.scanPort(p)
			results <- ScanResult{
				Port:   p,
				IsOpen: isOpen,
			}
		}(port)
	}

	wg.Wait()
}

// intentar establecer una conexion TCP con el target:puerto
func (s *TCPConnectScanner) scanPort(port int) bool {
	address := net.JoinHostPort(s.Target, fmt.Sprintf("%d", port)) //endpoint TCP estandar
	conn, err := net.DialTimeout("tcp", address, s.Timeout)

	if err != nil {
		return false
	}

	conn.Close()
	return true
}
