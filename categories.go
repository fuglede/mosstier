package main

import (
	"encoding/json"
	"errors"
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

func getAllCategories() (allCategories []category) {
	for _, class := range readCategories() {
		allCategories = append(allCategories, class.Categories...)
	} 
	return
}

func getMainCategories() []category {
	return readCategories()[0].Categories
}

func getChallengeCategories() []category {
	return readCategories()[1].Categories
}

func getCategoryByAbbr(abbr string) (category, error) {
	for _, cat := range getAllCategories() {
		if cat.Abbr == abbr {
			return cat, nil
		}
	}
	return category{}, errors.New("No such category")
}

func getCategoryById(id int) (category, error) {
	for _, cat := range getAllCategories() {
		if cat.Id == id {
			return cat, nil
		}
	}
	return category{}, errors.New("No such category")
}