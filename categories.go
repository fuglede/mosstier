package main

import (
	"encoding/json"
	"io/ioutil"
)


type allCategories struct {
	CategoryClasses	[]categoryClass `json:"categoryClasses"`
}

type categoryClass struct {
	Class		string		`json:"class"`
	Categories	[]category	`json:"categories"`
}

type category struct {
	Id			int		`json:"id"`
	Name		string	`json:"name"`
	Goal		string	`json:"goal"`
	Abbr		string	`json:"abbr"`
	Definition	string	`json:"definition"`
}

// readCategories returns a slice of all the news entries stored in data/news.json
func readCategories() []categoryClass {
	categoriesFile, _ := ioutil.ReadFile("data/categories.json")
	var allCategories allCategories
	json.Unmarshal(categoriesFile, &allCategories)
	return allCategories.CategoryClasses
}

func getMainCategories() []category {
	return readCategories()[0].Categories
}

func getChallengeCategories() []category {
	return readCategories()[1].Categories
}