package handlers

import (
	"context"
	"fmt"
	"go-scanner/cmd/go-scanner-web/app/views"
	"go-scanner/internal/app/scan"
	"go-scanner/internal/scanner"
	"go-scanner/internal/utils"
	"net/http"
	"time"
)

// struct de datos para renderizar la pagina
type PageData struct {
	Target  string
	Results []scanner.ScanResult
	Error   string
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
	h.renderer.Render(w, "scan.html", PageData{})
}

// procesar escaneo
func (h *Handler) Scan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawTarget := r.FormValue("target")
	data := PageData{Target: rawTarget}

	// validacion de input
	targets, err := utils.ParseTarget(rawTarget)
	if err != nil {
		data.Error = fmt.Sprintf("Invalid target: %v", err)
		h.renderer.Render(w, "scan.html", data)
		return
	}

	if len(targets) == 0 {
		data.Error = "No valid targets found"
		h.renderer.Render(w, "scan.html", data)
		return
	}

	// configuracion del escaneo via Service
	req := scan.ScanRequest{
		Targets:     targets,
		Ports:       "80,443,22,21,25,8080,3000", //default hardcodeada para (ESTO DEBE SER TEMPORAL WEON, ES POR VELOCIDAD, QUE NO SE TE OLVIDE, ESTABLCER SELECCION Y RANGO DE PORTS A ESCANEAR)
		ProfileName: "default",
		Options: scan.ScanOptions{
			Banner: true,
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

	data.Results = report.Results

	// renderizado
	err = h.renderer.Render(w, "scan.html", data)
	if err != nil {
		fmt.Printf("Template execution error: %v\n", err)
	}
}
