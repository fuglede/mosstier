package main

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
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
	query := "SELECT id, username, email, country, spelunker, steam, psn, xbla, twitch, youtube, freetext FROM users " + constraints
	statement, err := db.Prepare(query)
	if err != nil {
		return
	}
	defer statement.Close()
	err = statement.QueryRow(values...).Scan(&r.ID, &r.Username, &r.Email, &r.Country, &r.Spelunker, &r.Steam, &r.Psn, &r.Xbla, &r.Twitch, &r.YouTube, &r.FreeText)
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

// updatePassword sets a new password for the runner.
func (r *runner) updatePassword(password string) error {
	if r.Email == "" {
		return errors.New("user has no email set; it would be impossible to notify them")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	query, err := db.Prepare("UPDATE users SET pass = ? WHERE username = ?")
	if err != nil {
		return err
	}
	_, err = query.Exec(string(hashedPassword), r.Username)
	return err
}

// sendMail sends an email to the runner with a given subject and message body
func (r *runner) sendMail(subject string, body string) error {
	if r.Email == "" {
		return errors.New("user has no associated email address")
	}
	err := sendMail(r.Email, subject, body)
	return err
}
