package server

import (
	"fmt"
	"net/http"
	"time"
)

// inicializar y correr el servidor HTTP
func Start(port int, templateDir string) error {
	mux, err := SetupRoutes(templateDir)
	if err != nil {
		return fmt.Errorf("failed to setup routes: %w", err)
	}
	addr := fmt.Sprintf(":%d", port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 60 * time.Second, //tiempo para el scan
		IdleTimeout:  120 * time.Second,
	}

	fmt.Printf("starting web server on http://localhost%s\n", addr) //para acceder, con ctrl + click we, lo de siempre :v
	return srv.ListenAndServe()
}
