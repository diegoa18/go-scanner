package views

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
)

/*(TODO ESTO ES TEMPORAL, ES PARA UNA REPRESENTACION WEB INICIAL)
SE DEBE DE CAMBIAR Y PROPONER TODO UN APARTADO FRONTEND PARA LA CONTRUCCION DE UI/UX
POR AHORA LA UNICA TEMPLATE ES EL POBRE HTML QUE SE TANKEA TODO*/

// maneja la carga y ejecucion de templates
type Renderer struct {
	templateDir string
	templates   *template.Template
}

// nueva instancia para precargar templates
func NewRenderer(templateDir string) (*Renderer, error) {
	//busqueda de html's
	pattern := filepath.Join(templateDir, "*.html")

	//parsear los templates
	tmpl, err := template.ParseGlob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates in %s: %w", pattern, err)
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
