package tcp

//TCP CONNECT SCAN
import (
	"fmt"
	"go-scanner/internal/model"
	"go-scanner/internal/scanner"
	"go-scanner/internal/scanner/banner"
	"net"  //API de red
	"sync" //sincronizacion
	"time"
)

// encapsula todo el estado necesario para realizar un escaneo TCP
type TCPConnectScanner struct {
	Target       string              //host o ip objetivo
	Ports        []int               //puertos a escanear
	Timeout      time.Duration       //timeout por conexion
	Concurrency  int                 //numero maximo de conexiones concurrentes
	EnableBanner bool                //habilitar banner grabbing pasivo
	Metadata     *model.HostMetadata //contexto del descubrimiento
}

// nueva instacia de TCPConnectScanner
func NewTCPConnectScanner(target string, ports []int, timeout time.Duration, concurrency int, enableBanner bool, meta *model.HostMetadata) *TCPConnectScanner {
	return &TCPConnectScanner{
		Target:       target,
		Ports:        ports,
		Timeout:      timeout,
		Concurrency:  concurrency,
		EnableBanner: enableBanner, //banner grabbing
		Metadata:     meta,
	}
}

// debe iterar sobre los puertos y lanzar gorutinas limitadas
func (s *TCPConnectScanner) Scan(results chan<- scanner.ScanResult) {
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

			isOpen, bannerText := s.scanPort(p)
			state := scanner.PortStateClosed

			if isOpen {
				state = scanner.PortStateOpen
			}

			results <- scanner.ScanResult{
				Host:     s.Target,
				Port:     p,
				State:    state,
				Banner:   bannerText, //incluir banner grabbing
				Metadata: s.Metadata,
			}
		}(port)
	}

	wg.Wait()
}

// intentar establecer una conexion TCP con el target:puerto
func (s *TCPConnectScanner) scanPort(port int) (bool, string) {
	address := net.JoinHostPort(s.Target, fmt.Sprintf("%d", port)) //endpoint TCP estandar
	conn, err := net.DialTimeout("tcp", address, s.Timeout)

	if err != nil {
		return false, ""
	}

	defer conn.Close()

	var collectedBanner string
	if s.EnableBanner {
		collectedBanner, _ = banner.Grab(conn, port) //intentar obtener el banner
	}

	return true, collectedBanner
}
