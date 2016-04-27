package main

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
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

func exportOverviewHandler(w http.ResponseWriter, r *http.Request) {
	renderContent("tmpl/export.html", w, getAllCategories())
}

// isLegitExportFormat determines if a given format is one we know how to export
func isLegitExportFormat(format string) bool {
	legitFormats := [3]string{"csv", "json", "xml"}
	for _, legit := range legitFormats {
		if (legit == format) {
			return true
		}
	}
	return false
}

func exportWrHandler(w http.ResponseWriter, r *http.Request) {
	var exportFormat = mux.Vars(r)["exportFormat"]
	if (!isLegitExportFormat(exportFormat)) {
		http.NotFound(w, r)
	}
	worldRecords, _ := getAllWorldRecords()
	switch exportFormat {
	case "csv":
		// Set up the data
		w.Header().Set("Content-Type", "text/csv")
		body := make([][]string, len(worldRecords)+1)
		body[0] = []string{"Category", "Player", "Score/time", "Video link", "Comment"}
		for i, record := range worldRecords {
			body[i+1] = []string{record.Category.Name, record.Runner.Username, record.FormatScore(), record.Link, record.Comment}
		}
		// Output the data
		wr := csv.NewWriter(w)
		wr.Comma = ';'
		err := wr.WriteAll(body)
		if err != nil {
			log.Println("Could not write csv: ", err)
			http.Error(w, "Internal server error", 500)
		}
	case "xml":
		type recordXml struct {
			XMLName  xml.Name `xml:"record"`
			Category string `xml:"category,attr"`
			Player string `xml:"player"`
			Result string `xml:"result"`
			VideoLink string `xml:"videoLink"`
			Comment string `xml:"comment"`
		}
		type worldRecordsXml struct {
			XMLName xml.Name    `xml:"worldRecords"`
			Records []recordXml `xml:"record"`
		}	
		wrs := &worldRecordsXml{}
		for _, record := range worldRecords {
			newRecord := &recordXml{Category: record.Category.Name}
			newRecord.Player = record.Runner.Username
			newRecord.Result = record.FormatScore()
			newRecord.VideoLink = record.Link
			newRecord.Comment = record.Comment
			wrs.Records = append(wrs.Records, *newRecord)
		}
		body, err := xml.Marshal(wrs)
		if err != nil {
			log.Println("Could not write xml: ", err)
			http.Error(w, "Internal server error", 500)
		}
		w.Write(body)
	}
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

	type categoryData struct {
		Category category
		Runs     []run
	}
	data := categoryData{cat, runs}
	renderContent("tmpl/category.html", w, data)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	profileId, err := strconv.Atoi(vars["profileId"])
	if err != nil {
		log.Println("Could not parse profile id: ", err)
		http.NotFound(w, r)
		return
	}
	thisRunner, err := getRunnerById(profileId)
	if err != nil {
		log.Println("Could not find runner: ", err)
		http.NotFound(w, r)
		return
	}
	runs, err := getRunsByRunnerId(profileId)
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

func initializeHandlers() {
	staticHandler := http.FileServer(http.Dir("tmpl"))
	http.Handle("/css/", staticHandler)
	http.Handle("/font/", staticHandler)
	http.Handle("/img/", staticHandler)

	router := mux.NewRouter()
	router.HandleFunc("/", frontPageHandler)
	router.HandleFunc("/about", aboutHandler)
	router.HandleFunc("/export", exportOverviewHandler)
	router.HandleFunc("/export/all/{exportFormat:[a-z]+}", exportWrHandler)
	router.HandleFunc("/rules", rulesHandler)
	router.HandleFunc("/category/{categoryName:[a-z]+}", categoryHandler)
	router.HandleFunc("/profile/{profileId:[0-9]+}", profileHandler)
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
