package utils

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

// RenderTemplate renders a template with the given data.
// It assumes templates are located in the "templates" directory.
func RenderTemplate(w http.ResponseWriter, tmplName string, data interface{}) {
	tmplPath := filepath.Join("templates", tmplName)

	// Create a new template and parse the file
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Printf("Error checking template %s: %v", tmplName, err)
		http.Error(w, "Template Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Error executing template %s: %v", tmplName, err)
	}
}

// RenderWithLayout renders a template wrapped in a layout.
func RenderWithLayout(w http.ResponseWriter, tmplName string, data interface{}) {
	tmpl, err := template.ParseFiles("templates/layout.html", "templates/"+tmplName)
	if err != nil {
		log.Printf("Error checking template %s: %v", tmplName, err)
		http.Error(w, "Template Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		log.Printf("Error executing template %s: %v", tmplName, err)
	}
}
