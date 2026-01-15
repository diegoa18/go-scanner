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
	patterns := []string{
		filepath.Join(templateDir, "layout", "*.html"),
		filepath.Join(templateDir, "scan", "*.html"),
		filepath.Join(templateDir, "partials", "*.html"),
	}

	tmpl := template.New("") //tamplate vacio

	for _, pattern := range patterns {
		_, err := tmpl.ParseGlob(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to parse templates in %s: %w", pattern, err)
		}
	}

	return &Renderer{
		templateDir: templateDir,
		templates:   tmpl,
	}, nil
}

// renderiza un template especifico por nombre
func (r *Renderer) Render(w io.Writer, templateName string, data interface{}) error {
	return r.templates.ExecuteTemplate(w, templateName, data)
}
