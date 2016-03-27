package main

import (
	"encoding/json"
	"io/ioutil"
)

// readSpelunkerNames returns a slice containing the names of all spelunkers
func readSpelunkerNames() (spelunkers []string) {
	spelunkerFile, _ := ioutil.ReadFile("data/spelunkers.json")
	json.Unmarshal(spelunkerFile, &spelunkers)
	return
}