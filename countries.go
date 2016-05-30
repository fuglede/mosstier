package main

import (
	"encoding/json"
	"io/ioutil"
)

type countryJSON struct {
	Countries map[string]string
}

var countries map[string]string

// readCountryData initializes the map of country abbreviations
// from data/countries.json
func readCountries() {
	if countries != nil {
		return
	}
	countryFile, _ := ioutil.ReadFile("data/countries.json")
	json.Unmarshal(countryFile, &countries)
}
