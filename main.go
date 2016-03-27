package main

import (
	"html/template"
	"log"
	"net/http"
)

func contentHandler(t *template.Template, w http.ResponseWriter, r *http.Request) {
	t.ParseFiles("tmpl/base.html")
	t.ExecuteTemplate(w, "base", nil)
	t.Execute(w, nil)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("tmpl/about.html")
	contentHandler(t, w, r)
}

func frontPageHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("tmpl/base.html",  "tmpl/frontpage.html")
	news := readNews()
	err := t.ExecuteTemplate(w, "base", &news)
	if err != nil {
		log.Fatal(err)
	}
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
