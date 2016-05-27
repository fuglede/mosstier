package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

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
func renderContent(t string, w http.ResponseWriter, data interface{}) {
	// Besides whatever page specific content we have (given in `data`), we always want to render
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

func contactHandler(w http.ResponseWriter, r *http.Request) {
	// We will record any form errors in a single string.
	type contactData struct {
		MailSent bool
		Error    string
	}
	var mailSent = false
	var errorString string
	if r.Method == "POST" {
		success := true
		err := r.ParseForm()
		if err != nil {
			errorString += "Could not read form contents. "
			success = false
		}
		if len(r.Form["name"][0]) == 0 {
			errorString += "Name field can not be empty. "
			success = false
		}
		if len(r.Form["subject"][0]) == 0 {
			errorString += "Subject field can not be empty. "
			success = false
		}
		if len(r.Form["message"][0]) == 0 {
			errorString += "Message field can not be empty. "
			success = false
		}
		if success {
			subject := "Moss Tier contact form message: " + r.Form["subject"][0]
			message := "From: " + r.Form["name"][0] + "\r\n"
			if r.Form["email"] != nil {
				message += "Email: " + r.Form["email"][0] + "\r\n"
			}
			message += "\r\n\r\n" + r.Form["message"][0]
			err = sendMail(config.AdminEmail, subject, message)
			if err != nil {
				errorString = "Mail delivery failed."
				success = false
			} else {
				mailSent = true
			}
		}
	}
	renderContent("tmpl/contact.html", w, contactData{mailSent, errorString})
}

func frontPageHandler(w http.ResponseWriter, r *http.Request) {
	type frontPageData struct {
		News    []newsEntry
		Records []run
	}
	worldRecords, err := getAllWorldRecords()
	if err != nil {
		log.Println("Could not get world records: ", err)
		http.Error(w, "Internal server error", 500)
		return
	}
	data := frontPageData{readNews(), worldRecords}
	renderContent("tmpl/frontpage.html", w, data)
}

func rulesHandler(w http.ResponseWriter, r *http.Request) {
	renderContent("tmpl/rules.html", w, getAllCategories())
}

func categoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cat, err := getCategoryByAbbr(vars["categoryName"])
	if err != nil {
		http.NotFound(w, r)
		return
	}
	runs, err := getRunsByCategory(cat, 0)
	if err != nil {
		log.Println("Could not get runs: ", err)
		http.Error(w, "Internal server error", 500)
	}
	highlightedRunner := vars["runner"]
	type categoryData struct {
		Category          category
		Runs              []run
		HighlightedRunner string
	}
	data := categoryData{cat, runs, highlightedRunner}
	renderContent("tmpl/category.html", w, data)
}

func passwordResetHandler(w http.ResponseWriter, r *http.Request) {
	type passwordResetData struct {
		PasswordReset bool
		Error         string
	}
	passwordReset := false
	var errorString string

	if r.Method == "POST" {
		formValid := true
		err := r.ParseForm()
		if err != nil {
			errorString += "Could not parse form contents. "
			formValid = false
		}
		if len(r.Form["username"][0]) == 0 {
			errorString += "Username entry can not be empty. "
			formValid = false
		}
		if len(r.Form["email"][0]) == 0 {
			errorString += "Email entry can not be empty. "
			formValid = false
		}
		if formValid {
			user, err := getRunnerByUsernameAndEmail(r.Form["username"][0], r.Form["email"][0])
			if err != nil {
				errorString += "Could not find any user with that combination of username and email. "
			} else {
				newPassword := generatePassword()
				// Send a mail to the user with the password, before attempting to update it
				err = user.sendMail("Password reset", "Hi "+user.Username+". Someone (hopefully you) requested "+
					"a new password for you on Moss Tier. Here's your new one: "+newPassword)
				if err != nil {
					errorString += "Could not send you an email with your new password."
					log.Println(err)
				} else {
					err = user.updatePassword(generatePassword())
					if err != nil {
						errorString += "Could not update your password due to an unexpected error. "
						log.Println(err)
					} else {
						passwordReset = true
					}
				}
			}
		}
	}
	renderContent("tmpl/passwordreset.html", w, passwordResetData{passwordReset, errorString})
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	profileID, err := strconv.Atoi(vars["profileID"])
	if err != nil {
		log.Println("Could not parse profile id: ", err)
		http.NotFound(w, r)
		return
	}
	thisRunner, err := getRunnerByID(profileID)
	if err != nil {
		log.Println("Could not find runner: ", err)
		http.NotFound(w, r)
		return
	}
	runs, err := getRunsByRunnerID(profileID)
	if err != nil {
		log.Println("Could not get runs: ", err)
		http.Error(w, "Internal server error", 500)
	}
	type profileData struct {
		Runner runner
		Runs   []run
	}
	data := profileData{thisRunner, runs}
	renderContent("tmpl/profile.html", w, data)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	type registerData struct {
		Success bool
		Error   string
	}
	success := false
	errorString := ""
	data := registerData{success, errorString}
	renderContent("tmpl/register.html", w, data)
}

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
	router.HandleFunc("/password-reset", passwordResetHandler)
	router.HandleFunc("/profile/{profileID:[0-9]+}", profileHandler)
	router.HandleFunc("/register", registerHandler)
	router.HandleFunc("/rules", rulesHandler)
	http.Handle("/", router)
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

	err = http.ListenAndServe(fmt.Sprintf(":%d", config.WebserverPort), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
		db.Close()
	}
}
