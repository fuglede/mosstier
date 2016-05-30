package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
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

// Login sessions are handled using a single cookie store. Keys
// are updated on a per-server launch basis.
var cookieStore *sessions.CookieStore

func initializeCookieStore() {
	cookieStore = sessions.NewCookieStore(securecookie.GenerateRandomKey(64),
		securecookie.GenerateRandomKey(32))
	cookieStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}
}

// getActiveUser returns the currently logged in user, if any.
func getActiveUser(r *http.Request) (user runner, err error) {
	session, err := cookieStore.Get(r, "login")
	if err != nil {
		return
	}
	storedRunnerID, ok := session.Values["userID"]
	if !ok {
		err = errors.New("no user ID found in cookie")
		return
	}
	// Session values are map[string]interface{}, so we need to
	// cast and check the type.
	runnerID, ok := storedRunnerID.(int)
	if !ok {
		err = errors.New("user ID was not an integer")
		return
	}
	user, err = getRunnerByID(runnerID)
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
		PageContents        interface{}
	}
	user, _ := getActiveUser(r)
	templateDataVar := templateData{getMainCategories(), getChallengeCategories(), user, data}
	err := templates[t].ExecuteTemplate(w, "base", templateDataVar)
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
	router.HandleFunc("/export", exportOverviewHandler)
	router.HandleFunc("/export/all/{exportFormat:[a-z]+}", exportWrHandler)
	router.HandleFunc("/export/{categoryID:[0-9]+}/{exportFormat:[a-z]+}", exportCategoryHandler)
	router.HandleFunc("/login", loginHandler)
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

	initializeHandlers()

	err = http.ListenAndServe(fmt.Sprintf(":%d", config.WebserverPort), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
		db.Close()
	}
}
