package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type spelunker struct {
	ID   int
	Name string
}

var spelunkers []spelunker

// readSpelunkerNames read the names of all spelunkers into `spelunkers`
func readSpelunkerNames() {
	if spelunkers != nil {
		return
	}
	var spelunkerNames []string
	spelunkerFile, _ := ioutil.ReadFile("data/spelunkers.json")
	json.Unmarshal(spelunkerFile, &spelunkerNames)
	for id, name := range spelunkerNames {
		spelunker := spelunker{id, name}
		spelunkers = append(spelunkers, spelunker)
	}
	return
}

func getSpelunkerByID(id int) (spelunker, error) {
	for _, spelunker := range spelunkers {
		if spelunker.ID == id {
			return spelunker, nil
		}
	}
	return spelunker{}, errors.New("No spelunker with given id.")
}
