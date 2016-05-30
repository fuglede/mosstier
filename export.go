package main

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// isLegitExportFormat determines if a given format is one we know how to export
func isLegitExportFormat(format string) bool {
	legitFormats := [3]string{"csv", "json", "xml"}
	for _, legit := range legitFormats {
		if legit == format {
			return true
		}
	}
	return false
}

// exportOverviewHandler handles requests to /export
func exportOverviewHandler(w http.ResponseWriter, r *http.Request) {
	renderContent("tmpl/export.html", r, w, getAllCategories())
}

// exportWrHandler handles requests to /export/all/[a-z]+
func exportWrHandler(w http.ResponseWriter, r *http.Request) {
	var exportFormat = mux.Vars(r)["exportFormat"]
	if !isLegitExportFormat(exportFormat) {
		http.NotFound(w, r)
		return
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
	case "json":
		w.Header().Set("Content-Type", "application/json")
		type recordJson struct {
			Category  string `json:"category"`
			Player    string `json:"player"`
			Result    string `json:"result"`
			Videolink string `json:"videoLink"`
			Comment   string `json:"comment"`
		}
		type worldRecordsJson struct {
			WorldRecords []recordJson `json:"worldRecords"`
		}
		wrs := &worldRecordsJson{}
		for _, record := range worldRecords {
			wrs.WorldRecords = append(wrs.WorldRecords,
				recordJson{record.Category.Name,
					record.Runner.Username,
					record.FormatScore(),
					record.Link,
					record.Comment})
		}
		body, err := json.Marshal(wrs)
		if err != nil {
			log.Println("Could not write json: ", err)
			http.Error(w, "Internal server error", 500)
		}
		w.Write(body)
	case "xml":
		w.Header().Set("Content-Type", "text/xml")
		type recordXml struct {
			XMLName   xml.Name `xml:"record"`
			Category  string   `xml:"category,attr"`
			Player    string   `xml:"player"`
			Result    string   `xml:"result"`
			VideoLink string   `xml:"videoLink"`
			Comment   string   `xml:"comment"`
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

// exportCategoryHandler handles requests to /export/[0-9]+/[a-z]+
func exportCategoryHandler(w http.ResponseWriter, r *http.Request) {
	exportFormat := mux.Vars(r)["exportFormat"]
	if !isLegitExportFormat(exportFormat) {
		http.NotFound(w, r)
	}
	categoryID, _ := strconv.Atoi(mux.Vars(r)["categoryID"])
	category, err := getCategoryByID(categoryID)
	if err != nil {
		log.Println("Could not get category: ", err)
		http.NotFound(w, r)
		return
	}
	runs, err := getRunsByCategory(category, 0)
	if err != nil {
		log.Println("Could not get runs: ", err)
		http.Error(w, "Internal server error", 500)
	}
	// Enter unfortunate duplication. :(
	switch exportFormat {
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		body := make([][]string, len(runs)+1)
		body[0] = []string{"Rank", "Player", category.Goal, "Video link", "Comment"}
		for i, run := range runs {
			body[i+1] = []string{strconv.Itoa(i + 1), run.Runner.Username, run.FormatScore(), run.Link, run.Comment}
		}
		// Output the data
		wr := csv.NewWriter(w)
		wr.Comma = ';'
		err := wr.WriteAll(body)
		if err != nil {
			log.Println("Could not write csv: ", err)
			http.Error(w, "Internal server error", 500)
		}
	case "json":
		w.Header().Set("Content-Type", "application/json")
		type runJSON struct {
			Rank      int    `json:"rank"`
			Player    string `json:"player"`
			Result    string `json:"result"`
			Videolink string `json:"videoLink"`
			Comment   string `json:"comment"`
		}
		type runsJSON struct {
			Category string    `json:"category"`
			Runs     []runJSON `json:"runs"`
		}
		runsForExport := &runsJSON{}
		runsForExport.Category = category.Name
		for i, run := range runs {
			runsForExport.Runs = append(runsForExport.Runs,
				runJSON{i + 1, run.Runner.Username, run.FormatScore(), run.Link, run.Comment})
		}
		body, err := json.Marshal(runsForExport)
		if err != nil {
			log.Println("Could not write json: ", err)
			http.Error(w, "Internal server error", 500)
		}
		w.Write(body)
	case "xml":
		w.Header().Set("Content-Type", "text/xml")
		type runXML struct {
			XMLName   xml.Name `xml:"run"`
			Rank      int      `xml:"rank"`
			Player    string   `xml:"player"`
			Result    string   `xml:"result"`
			VideoLink string   `xml:"videoLink"`
			Comment   string   `xml:"comment"`
		}
		type runsXML struct {
			XMLName  xml.Name `xml:"leaderboards"`
			Category string   `xml:"category,attr"`
			Runs     []runXML `xml:"runs"`
		}
		runsForExport := &runsXML{Category: category.Name}
		for i, run := range runs {
			runForExport := &runXML{}
			runForExport.Rank = i + 1
			runForExport.Player = run.Runner.Username
			runForExport.Result = run.FormatScore()
			runForExport.VideoLink = run.Link
			runForExport.Comment = run.Comment
			runsForExport.Runs = append(runsForExport.Runs, *runForExport)
		}
		body, err := xml.Marshal(runsForExport)
		if err != nil {
			log.Println("Could not write xml: ", err)
			http.Error(w, "Internal server error", 500)
		}
		w.Write(body)
	}

}
