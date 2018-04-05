package main

import (
	"net/http"
	"path/filepath"
	"sync"

	"github.com/alecthomas/template"
)

// Template is used to serve HTTP templates
type Template struct {
	Files []string
	tpl   *template.Template
	once  sync.Once
}

// Execute shows the template on the provided ResponseWriter and passes the provided
// The template will be initialized once
func (t *Template) Execute(w http.ResponseWriter, data map[string]interface{}) {
	// Initialize the template
	t.once.Do(func() {
		// Create an array containing the files with a prepended path
		files := []string{}
		for _, file := range t.Files {
			files = append(files, filepath.Join("templates", file))
		}

		// Parse the files, pass the error to the ResponseWriter if an error occures
		tpl, err := template.ParseFiles(files...)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		// Store the parsed template
		t.tpl = tpl
	})

	// Show the template if t.tpl is valid
	if t.tpl != nil {
		t.tpl.ExecuteTemplate(w, "main", data)
	}
}
