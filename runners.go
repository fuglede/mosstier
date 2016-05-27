package main

import (
	"golang.org/x/crypto/bcrypt"
	"log"
)
type runner struct {
	ID             int
	Username       string
	Email          string
	Country        string
	Spelunker      int
	Steam          int
	Psn            string
	Xbla           string
	Twitch         string
	YouTube        string
	FreeText       string
	EmailFlag      int
	EmailWr        int
	EmailChallenge int
}

// searchRunner returns a user on the site, found by applying a given filter
func searchRunner(constraints string, values ...interface{}) (r runner, err error) {
	query := "SELECT id, username, country, spelunker, steam, psn, xbla, twitch, youtube, freetext FROM users " + constraints
	statement, err := db.Prepare(query)
	if err != nil {
		return
	}
	defer statement.Close()
	err = statement.QueryRow(values...).Scan(&r.ID, &r.Username, &r.Country, &r.Spelunker, &r.Steam, &r.Psn, &r.Xbla, &r.Twitch, &r.YouTube, &r.FreeText)
	return
}

// getRunnerById returns the user with the specific numerical id
func getRunnerByID(id int) (runner, error) {
	return searchRunner("WHERE id = ?", id)
}

// getRunnerByUsernameAndEmail returns the user with a given username and email
func getRunnerByUsernameAndEmail(username string, email string) (runner, error) {
	return searchRunner("WHERE username = ? AND email = ?", username, email)
}

// UpdatePassword sets a new password for the runner.
func (r *runner) UpdatePassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	query = "UPDATE users SET password = " + hashedPassword + " WHERE username = " + r.Username
	log.Println(string(hashedPassword))
	if err != nil {
		return err
	}
	return nil
}