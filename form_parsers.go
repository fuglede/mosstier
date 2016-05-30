package main

import (
	"errors"
	"log"
	"net/http"
)

// getFormValue returns the value of a given POST parameter if non-empty
func getFormValue(r *http.Request, key string) (string, error) {
	formValue, ok := r.Form[key]
	if ok {
		return formValue[0], nil
	}
	return "", errors.New(key + " was not a POST parameter")
}

// loginParser parses POST requests to "/login". Returns the
// user to log in on success.
func loginFormParser(r *http.Request) (runner, error) {
	err := r.ParseForm()
	if err != nil {
		return runner{}, errors.New("Could not parse form contents.")
	}
	username, err := getFormValue(r, "username")
	if err != nil || !isLegitUsername(username) {
		log.Println(username)
		return runner{}, errors.New("Invalid username.")
	}
	password, err := getFormValue(r, "password")
	if err != nil || !isLegitPassword(password) {
		return runner{}, errors.New("Invalid password.")
	}
	user, err := getRunnerByUsername(username)
	if err != nil {
		return runner{}, errors.New("Could not find user with that username.")
	}
	err = user.testLogin(password)
	if err != nil {
		return runner{}, errors.New("Incorrect password.")
	}
	return user, nil
}

// passwordResetHandler parses POST requests to "/password-reset",
// and returns the user whose password should be reset.
func passwordResetFormParser(r *http.Request) (runner, error) {
	err := r.ParseForm()
	if err != nil {
		return runner{}, errors.New("Could not parse form contents.")
	}
	username, err := getFormValue(r, "username")
	if err != nil || username == "" {
		return runner{}, errors.New("Username entry can not be empty.")
	}
	email, err := getFormValue(r, "email")
	if err != nil || email == "" {
		return runner{}, errors.New("Email entry can not be empty.")
	}
	user, err := getRunnerByUsernameAndEmail(username, email)
	if err != nil {
		return runner{}, errors.New("Could not find any user with that combination of username and email. ")
	}
	return user, nil
}
