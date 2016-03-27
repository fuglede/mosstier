package main

import (
	"encoding/json"
	"io/ioutil"
)


type news struct {
	NewsEntries	[]newsEntry `json:"news"`
}

type newsEntry struct {
	Date		string	`json:"date"`
	Contents	string	`json:"contents"`
}

// readNews returns a slice of all the news entries stored in data/news.json
func readNews() []newsEntry {
	newsFile, _ := ioutil.ReadFile("data/news.json")
	var allNews news
	json.Unmarshal(newsFile, &allNews)
	return allNews.NewsEntries
}