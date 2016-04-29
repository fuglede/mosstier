package main

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

// exportOverviewHandler handles requests to /export
func exportOverviewHandler(w http.ResponseWriter, r *http.Request) {
	renderContent("tmpl/export.html", w, getAllCategories())
}

// exportWrHandler handles requests to /export/all/[a-z]+
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
	case "json":
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
			wrs.WorldRecords = append(wrs.WorldRecords, recordJson{record.Category.Name, record.Runner.Username, record.FormatScore(), record.Link, record.Comment})
		}
		body, err := json.Marshal(wrs)
		if err != nil {
			log.Println("Could not write json: ", err)
			http.Error(w, "Internal server error", 500)
		}
		w.Write(body)
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