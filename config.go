package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type configType struct {
	DbConnection	string	`json:"dbConnection"`
}

var config configType

// readConfig reads the current configuration if it has not already been read.
func readConfig() {
	configFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal("no config file found. Create one as config.json")
	}
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatal("Could not read config file.")
	}
	return
}