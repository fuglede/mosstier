package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type configType struct {
	DbConnection  string `json:"dbConnection"`
	WebserverPort int    `json:"webserverPort"`
	AdminEmail    string `json:"adminEmail"`
	SMTPHost      string `json:"smtpHost"`
	SMTPPort      int    `json:"smtpPort"`
	SMTPUsername  string `json:"smtpUsername"`
	SMTPPassword  string `json:"smtpPassword"`
	MailSender    string `json:"mailSender"`
	Moderators    []int  `json:"moderatorIDs"`
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
