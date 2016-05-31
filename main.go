package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
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
// The control of the input data is handled by the handlers in handlers.go
func renderContent(t string, r *http.Request, w http.ResponseWriter, data interface{}) {
	// Besides whatever page specific content we have (given in `data`), we always want to render
	// a list of categories.
	type templateData struct {
		MainCategories      []category
		ChallengeCategories []category
		ActiveUser          runner
		UserLoggedIn        bool
		PageContents        interface{}
	}
	user, err := getActiveUser(r)
	loggedIn := err == nil
	templateDataVar := templateData{
		getMainCategories(),
		getChallengeCategories(),
		user,
		loggedIn,
		data}
	err = templates[t].ExecuteTemplate(w, "base", templateDataVar)
	if err != nil {
		log.Println(err)
	}
}

// initializeHandlers sets up the relevant handlers for all possible
// GET and POST requests.
func initializeHandlers() {
	staticHandler := http.FileServer(http.Dir("tmpl"))
	http.Handle("/css/", staticHandler)
	http.Handle("/font/", staticHandler)
	http.Handle("/img/", staticHandler)

	router := mux.NewRouter()
	router.HandleFunc("/", frontPageHandler)
	router.HandleFunc("/about", aboutHandler)
	router.HandleFunc("/category/{categoryName:[a-z]+}", categoryHandler)
	router.HandleFunc("/category/{categoryName:[a-z]+}/find/{runner:[0-9a-zA-Z_-]+}", categoryHandler)
	router.HandleFunc("/contact", contactHandler)
	router.HandleFunc("/edit-profile", editProfileHandler)
	router.HandleFunc("/export", exportOverviewHandler)
	router.HandleFunc("/export/all/{exportFormat:[a-z]+}", exportWrHandler)
	router.HandleFunc("/export/{categoryID:[0-9]+}/{exportFormat:[a-z]+}", exportCategoryHandler)
	router.HandleFunc("/login", loginHandler)
	router.HandleFunc("/log-out", logOutHandler)
	router.HandleFunc("/password-reset", passwordResetHandler)
	router.HandleFunc("/profile/{profileID:[0-9]+}", profileHandler)
	router.HandleFunc("/register", registerHandler)
	router.HandleFunc("/report/{runID:[0-9]+}", reportHandler)
	router.HandleFunc("/rules", rulesHandler)
	http.Handle("/", router)
}

func main() {
	err := initializeTemplates()
	initializeCookieStore()
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
	readCountries()

	initializeHandlers()

	err = http.ListenAndServe(fmt.Sprintf(":%d", config.WebserverPort), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
		db.Close()
	}
}
