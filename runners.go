package main

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type runner struct {
	ID             int
	Username       string
	Password       string
	Email          string
	Country        string
	Spelunker      spelunker
	Steam          int
	Psn            string
	Xbla           string
	Twitch         string
	YouTube        string
	FreeText       string
	EmailFlag      bool
	EmailWr        bool
	EmailChallenge bool
}

// searchRunner returns a user on the site, found by applying a given filter
func searchRunner(constraints string, values ...interface{}) (r runner, err error) {
	query := "SELECT id, username, pass, email, country, spelunker, steam, psn, xbla, twitch, youtube, freetext FROM users " + constraints
	statement, err := db.Prepare(query)
	if err != nil {
		return
	}
	defer statement.Close()
	var spelunkerID int
	err = statement.QueryRow(values...).Scan(&r.ID, &r.Username, &r.Password, &r.Email, &r.Country, &spelunkerID, &r.Steam, &r.Psn, &r.Xbla, &r.Twitch, &r.YouTube, &r.FreeText)
	r.Spelunker, _ = getSpelunkerByID(spelunkerID)
	return
}

// getRunnerById returns the user with the specific numerical id
func getRunnerByID(id int) (runner, error) {
	return searchRunner("WHERE id = ?", id)
}

// getRunnerByUsername returns the user with a given username
func getRunnerByUsername(username string) (runner, error) {
	return searchRunner("WHERE username = ?", username)
}

// getRunnerByUsernameAndEmail returns the user with a given username and email
func getRunnerByUsernameAndEmail(username, email string) (runner, error) {
	return searchRunner("WHERE username = ? AND email = ?", username, email)
}

// makeUser creates a new user with a given username, email, and password
func makeUser(username, email, password string) (err error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return
	}
	stmt, err := db.Prepare("INSERT INTO users SET username = ?, email = ?, pass = ?")
	if err != nil {
		return
	}
	_, err = stmt.Exec(username, email, string(hashedPassword))
	return
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

// testLogin tries to log in a user with a given password
func (r *runner) testLogin(password string) (err error) {
	// We are currently deprecating passwords that begin with $2y$
	hashedPassword := strings.Replace(r.Password, "$2y$", "$2a$", 1)
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return
	}
	// At this point, let the user in, but update their password to
	// the new form if it's not already.
	if r.Password[0:4] == "$2y$" {
		// No biggy if this doesn't work, so we ignore errors for once.
		r.updatePassword(password)
	}
	return
}

// sendMail sends an email to the runner with a given subject and message body
func (r *runner) sendMail(subject, body string) error {
	if r.Email == "" {
		return errors.New("user has no associated email address")
	}
	err := sendMail(r.Email, subject, body)
	return err
}

// isModerator returns true iff the user is a site moderator
func (r *runner) isModerator() bool {
	for _, moderatorID := range config.Moderators {
		if moderatorID == r.ID {
			return true
		}
	}
	return false
}

// formatCountry produces the full name of the runner's chosen country
func (r *runner) FormatCountry() string {
	return countries[r.Country]
}
