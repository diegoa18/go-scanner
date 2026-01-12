package handlers

import (
	"context"
	"fmt"
	"go-scanner/cmd/go-scanner-web/app/views"
	"go-scanner/internal/app/scan"
	"net/http"
	"time"
)

// struct de datos para renderizar la pagina
type PageData struct {
	Title  string
	Target string
	Report *scan.ScanReport
	Error  string
}

// handler agrupa las dependencias del handler de escaneo
type Handler struct {
	renderer *views.Renderer
}

// nueva instancia del handler
func NewHandler(renderer *views.Renderer) *Handler {
	return &Handler{
		renderer: renderer,
	}
}

// renderizacion de la pagina HOME
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	h.renderer.Render(w, "page", PageData{Title: "Go-Scanner | Home"})
}

// procesar escaneo
func (h *Handler) Scan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawTarget := r.FormValue("target")
	data := PageData{
		Title:  "Go-Scanner | Results",
		Target: rawTarget,
	}

	// configuracion del escaneo asumiendo input crudo
	req := scan.ScanRequest{
		Targets:     []string{r.FormValue("target")},
		Ports:       r.FormValue("ports"),
		ProfileName: r.FormValue("profile"),
		Options: scan.ScanOptions{
			Banner: r.FormValue("banner") == "true",
			Probe:  r.FormValue("probe") == "true",
			// ProbeTypes -> empty; para usar defaults del profile/cli logic
		},
	}

	//ruuuun
	svc := scan.NewService()

	// contexto con un timeout para la web
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	report, err := svc.Run(ctx, req)
	if err != nil {
		data.Error = fmt.Sprintf("Scan failed: %v", err)
		h.renderer.Render(w, "scan.html", data)
		return
	}

	// transparencia (reporte)
	data.Report = report

	// renderizado
	err = h.renderer.Render(w, "page", data)
	if err != nil {
		fmt.Printf("Template execution error: %v\n", err)
	}
}
