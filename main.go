package main

import (
	"html/template"
	"log"
	"net/http"
)

func contentHandler(templateFile string, w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("tmpl/base.html", templateFile)
	t.ExecuteTemplate(w, "base", nil)
	t.Execute(w, nil)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	contentHandler("tmpl/about.html", w, r)
}

func frontPageHandler(w http.ResponseWriter, r *http.Request) {
	contentHandler("tmpl/frontpage.html", w, r)
}

func main() {
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
