package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var templates = map[string]*template.Template{}

func loadTemplates() error {
	entries, err := os.ReadDir("templates")
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		"formatDate":       formatDate,
		"inputDate":        inputDate,
		"joinTags":         joinTags,
		"uniqueCategories": uniqueCategories,
		"uniqueTags":       uniqueTags,
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if name == "layout.html" {
			continue
		}

		page := strings.TrimSuffix(name, filepath.Ext(name))

		templates[page] = template.Must(template.New("").
			Funcs(funcMap).
			ParseFiles(
				"templates/layout.html",
				filepath.Join("templates", name),
			),
		)

	}

	return nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	t, ok := templates[tmpl]
	if !ok {
		http.Error(w, "template not found", http.StatusInternalServerError)
		return
	}

	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func render404(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	renderTemplate(w, "404", nil)
}

func render500(w http.ResponseWriter, err error) {
	log.Println(err)

	w.WriteHeader(http.StatusInternalServerError)
	renderTemplate(w, "500", nil)
}
