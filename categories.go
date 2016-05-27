package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type allCategories struct {
	CategoryClasses []categoryClass `json:"categoryClasses"`
}

type categoryClass struct {
	Class      string     `json:"class"`
	Categories []category `json:"categories"`
}

type category struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Goal       string `json:"goal"`
	Abbr       string `json:"abbr"`
	Definition string `json:"definition"`
}

// readCategories returns a slice of all the categories stored in data/categories.json,
// split up into their respective classes
func readCategories() []categoryClass {
	categoriesFile, _ := ioutil.ReadFile("data/categories.json")
	var allCategories allCategories
	json.Unmarshal(categoriesFile, &allCategories)
	return allCategories.CategoryClasses
}

// getAllCategories returns a slice of all categories
func getAllCategories() (allCategories []category) {
	for _, class := range readCategories() {
		allCategories = append(allCategories, class.Categories...)
	}
	return
}

// getMainCategories returns a slice of all the categories considered
// to be the most interesting ones.
func getMainCategories() []category {
	return readCategories()[0].Categories
}

// isMain returns true iff the given category is a main category.
func (cat *category) isMain() bool {
	for _, mainCat := range getMainCategories() {
		if mainCat.ID == cat.ID {
			return true
		}
	}
	return false
}

// getMainCategories returns a slice of all non-main categories.
func getChallengeCategories() []category {
	return readCategories()[1].Categories
}

// getCategoryByAbbr returns the category with a given abbreviation.
func getCategoryByAbbr(abbr string) (category, error) {
	for _, cat := range getAllCategories() {
		if cat.Abbr == abbr {
			return cat, nil
		}
	}
	return category{}, errors.New("No such category")
}

// getCategoryByID returns the category with a given integer ID.
func getCategoryByID(id int) (category, error) {
	for _, cat := range getAllCategories() {
		if cat.ID == id {
			return cat, nil
		}
	}
	return category{}, errors.New("No such category")
}
