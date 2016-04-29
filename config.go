package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type configType struct {
	DbConnection  string `json:"dbConnection"`
	WebserverPort int    `json:"webserverPort"`
	SmtpHost      string `json:"smtpHost"`
	SmtpPort      int    `json:"smtpPort"`
	SmtpUsername  string `json:"smtpUsername"`
	// Maybe this should not sit in memory ...
	SmtpPassword  string `json:"smtpPassword"`
	MailSender    string `json:"mailSender"`
}

var config configType

// readConfig reads the current configuration if it has not already been read.
func readConfig() (err error) {
	configFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		return errors.New("no config file found. Create one as config.json")
	}
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return errors.New("Could not read config file.")
	}
	return
}
