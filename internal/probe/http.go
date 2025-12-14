package probe

import (
	"fmt"
	"net/http" //cliente http
	"strings"
	"time"
)

// interfaz Prober para HTTP/HTTPS
type HTTPProbe struct{}

// nueva instancia de este
func NewHTTPProbe() *HTTPProbe {
	return &HTTPProbe{}
}

// ejecuta un request ligero HTTP/HTTPS
func (p *HTTPProbe) Probe(target string, port int, timeout time.Duration) (string, error) {
	//determinar schema
	scheme := "http"
	if port == 443 || port == 8443 {
		scheme = "https"
	}

	url := fmt.Sprintf("%s://%s:%d", scheme, target, port)

	//cliente HTTP con timeout estricto
	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse //no seguir redirects
		},
	}

	//primero el HEAD, luego mediante GET solo para el body
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	//construir el banner
	var info []string

	info = append(info, fmt.Sprintf("[%s]", resp.Status))

	//headers relevantes
	if server := resp.Header.Get("Server"); server != "" {
		info = append(info, fmt.Sprintf("Server: %s", server))
	}
	if powered := resp.Header.Get("X-Powered-By"); powered != "" {
		info = append(info, fmt.Sprintf("X-Powered-By: %s", powered))
	}

	//si no hay headers relevantes, tendria que ver otra forma, pero debo tomar onche calmao noma
	//evitare leer el body de manera innecesaria

	return strings.Join(info, " | "), nil
}
