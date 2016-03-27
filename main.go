package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

// We store all templates on first launch for efficiency.
var templates map[string]*template.Template

func initializeTemplates() {
    if templates == nil {
        templates = make(map[string]*template.Template)
    }
    templateFiles, err := filepath.Glob("tmpl/*.html")
    if err != nil {
        log.Fatal(err)
    }
    for _, t := range templateFiles {
    	if t != "tmpl/base.html" {
        	templates[t] = template.Must(template.ParseFiles("tmpl/base.html", t))
    	}
    }
}

// renderContent parses the content (given as a template) and puts it into our base template. 
func renderContent(t string, w http.ResponseWriter, data interface{}) {
	templates[t].ExecuteTemplate(w, "base", data)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	renderContent("tmpl/about.html", w, nil)
}

func frontPageHandler(w http.ResponseWriter, r *http.Request) {
	renderContent("tmpl/frontpage.html", w, readNews())
}

func main() {
	initializeTemplates()
	
	staticHandler := http.FileServer(http.Dir("tmpl"))
	http.Handle("/css/", staticHandler)
	http.Handle("/font/", staticHandler)
	http.Handle("/img/", staticHandler)
	
	http.HandleFunc("/", frontPageHandler)
	http.HandleFunc("/about", aboutHandler)
	
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
