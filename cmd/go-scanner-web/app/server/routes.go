package server

import (
	"go-scanner/cmd/go-scanner-web/app/handlers"
	"go-scanner/cmd/go-scanner-web/app/views"
	"net/http"
)

// configura las rutas de la aplicacion (que no son muchas por ahora)
func SetupRoutes(templateDir string) (*http.ServeMux, error) {
	// inicializar dependencias
	renderer, err := views.NewRenderer(templateDir)
	if err != nil {
		return nil, err
	}

	h := handlers.NewHandler(renderer)

	mux := http.NewServeMux()

	// servir archivos estaticos (css, js)
	fs := http.FileServer(http.Dir("cmd/go-scanner-web/app/views/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/", h.Home)
	mux.HandleFunc("/scan", h.Scan)
	return mux, nil
}
