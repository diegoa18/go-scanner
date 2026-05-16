package probe

//REGISTRO DE PROBERS
import (
	"go-scanner/internal/scanner/probe/http"
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

// default probers
func init() {
	//http/s apuntan al mismo prober
	h := http.NewHTTPProbe()
	Register("http", h)
	Register("https", h)
}


