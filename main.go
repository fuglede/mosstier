package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

// We store all templates on first launch for efficiency.
var templates map[string]*template.Template

// initializeTemplates populates `templates` for use in our handlers. The logic is that each
// of our templates is composed by base.html and some other HTML template.
func initializeTemplates() (err error) {
	if templates == nil {
		templates = make(map[string]*template.Template)
	}
	templateFiles, err := filepath.Glob("tmpl/*.html")
	if err != nil {
		return
	}
	for _, t := range templateFiles {
		if t != "tmpl/base.html" {
			templates[t] = template.Must(template.ParseFiles("tmpl/base.html", t))
		}
	}
	return
}

// renderContent parses the content (given as a template) and puts it into our base template.
func renderContent(t string, w http.ResponseWriter, data interface{}) {
	// Besides whatever page specific content we have, we always want to render
	// a list of categories.
	type templateData struct {
		MainCategories      []category
		ChallengeCategories []category
		PageContents        interface{}
	}
	templateDataVar := templateData{getMainCategories(), getChallengeCategories(), data}
	err := templates[t].ExecuteTemplate(w, "base", templateDataVar)
	if err != nil {
		log.Println(err)
	}
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	renderContent("tmpl/about.html", w, nil)
}

func frontPageHandler(w http.ResponseWriter, r *http.Request) {
	renderContent("tmpl/frontpage.html", w, readNews())
}

func rulesHandler(w http.ResponseWriter, r *http.Request) {
	renderContent("tmpl/rules.html", w, getAllCategories())
}

func categoryHandler(w http.ResponseWriter, r *http.Request) {
	categoryString := r.URL.Path[10:]
	cat, err := getCategoryByAbbr(categoryString)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	type runWithRank struct {
		Run  run
		Rank int
	}
	type categoryData struct {
		Category category
		Runs     []run
	}
	runs, err := getRunsByCategory(cat)
	if err != nil {
		log.Println("Could not get runs: ", err)
		http.Error(w, err.Error(), 500)
	}

	data := categoryData{cat, runs}
	renderContent("tmpl/category.html", w, data)
}

func initializeHandlers() {
	staticHandler := http.FileServer(http.Dir("tmpl"))
	http.Handle("/css/", staticHandler)
	http.Handle("/font/", staticHandler)
	http.Handle("/img/", staticHandler)

	http.HandleFunc("/", frontPageHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/rules", rulesHandler)
	http.HandleFunc("/category/", categoryHandler)
}

func main() {
	err := initializeTemplates()
	if err != nil {
		log.Fatal("Could not initialise templates: ", err)
	}
	err = readConfig()
	if err != nil {
		log.Fatal(err)
	}
	err = initializeDatabase()
	if err != nil {
		log.Fatal("Could not initialise database: ", err)
	}
	readSpelunkerNames()

	initializeHandlers()

	err = http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
		db.Close()
	}
}
