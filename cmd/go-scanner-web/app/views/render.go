package views

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
)

//ya no se lo tankea todo scan.html, PERO, por ahora se emplea alta segmentacion de HTML ya que se renderiza con Go

// maneja la carga y ejecucion de templates
type Renderer struct {
	templateDir string
	templates   *template.Template
}

// nueva instancia para precargar templates
func NewRenderer(templateDir string) (*Renderer, error) {
	// Definir patrones de búsqueda para estructura modular
	patterns := []string{
		filepath.Join(templateDir, "layout", "*.html"),
		filepath.Join(templateDir, "scan", "*.html"),
		filepath.Join(templateDir, "partials", "*.html"),
	}

	// Crear template base vacio
	tmpl := template.New("")

	// Parsear cada patron
	for _, pattern := range patterns {
		_, err := tmpl.ParseGlob(pattern)
		if err != nil {
			// Es aceptable que un directorio este vacio o no exista al inicio?
			// Para esta refactorizacion estricta, si falla es error.
			return nil, fmt.Errorf("failed to parse templates in %s: %w", pattern, err)
		}
	}

	return &Renderer{
		templateDir: templateDir,
		templates:   tmpl,
	}, nil
}

// renderiza un template especifico por nombre
// OJO: Ahora el nombre del template suele ser "base" o "page" que definen bloques
func (r *Renderer) Render(w io.Writer, templateName string, data interface{}) error {
	// IMPORTANTE: Al usar layouts, ejecutamos el template que define la estructura completa
	// En nuestro caso, 'page.html' define "page" que invoca a "base".
	// Si templateName es "scan/page.html" (file path), necesitamos ejecutar el bloque que define.
	// Por convención simple, usaremos el nombre del define.
	return r.templates.ExecuteTemplate(w, templateName, data)
}
