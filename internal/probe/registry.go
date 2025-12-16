package probe

//REGISTRO DE PROBERS
import (
	"fmt"
	"go-scanner/internal/probe/http"
	"strings"
)

// registro global de probers
var registry = make(map[string]Prober)

// registrar prober
func Register(name string, p Prober) {
	registry[strings.ToLower(name)] = p
}

// obtener prober mediante nombre
func Get(name string) (Prober, bool) {
	p, ok := registry[strings.ToLower(name)]
	return p, ok
}

// probers disponibles
func Available() []string {
	keys := make([]string, 0, len(registry))
	for k := range registry {
		keys = append(keys, k)
	}
	return keys
}

// default probers
func init() {
	//http/s apuntan al mismo prober
	h := http.NewHTTPProbe()
	Register("http", h)
	Register("https", h)
}

// para obtener mas de un prober a la vez
func GetProbers(types []string) (map[string]Prober, error) {
	probers := make(map[string]Prober)
	for _, t := range types {
		//soporte para "all"
		if t == "all" {
			return registry, nil
		}

		if p, ok := Get(t); ok {
			probers[t] = p
		} else {
			return nil, fmt.Errorf("probe type not found: %s", t)
		}
	}
	return probers, nil
}
