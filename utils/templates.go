package utils

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

// RenderTemplate merender template HTML dengan data yang diberikan.
// Fungsi ini berasumsi file template berada di direktori "templates".
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

// RenderWithLayout merender template yang dibungkus dalam layout utama.
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
