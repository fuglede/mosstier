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

// Below follows parsers for all forms on the websit, ordered alphabetically.

// contactFormParser parses the contact form, and returns the name, email,
// subject, and message on success.
func contactFormParser(r *http.Request) (string, string, string, string, error) {
	err := r.ParseForm()
	if err != nil {
		return "", "", "", "", errors.New("Could not read form contents.")
	}
	name, err := getFormValue(r, "name")
	if err != nil || name == "" {
		return "", "", "", "", errors.New("Name field can not be empty.")
	}
	email, err := getFormValue(r, "email")
	if err != nil {
		return "", "", "", "", errors.New("Could not parse email.")
	}
	subject, err := getFormValue(r, "subject")
	if err != nil || subject == "" {
		return "", "", "", "", errors.New("Subject field can not be empty.")
	}
	message, err := getFormValue(r, "message")
	if err != nil || message == "" {
		return "", "", "", "", errors.New("Message field can not be empty.")
	}
	return name, email, subject, message, nil
}

// loginFormParser parses POST requests to "/login". Returns the
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

// passwordResetFormHandler parses POST requests to "/password-reset",
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

// registerFormParser parses forms posted to "/register" and returns
// the username, email, and password posted on success.
func registerFormParser(r *http.Request) (string, string, string, error) {
	err := r.ParseForm()
	if err != nil {
		return "", "", "", errors.New("Could not parse form contents.")
	}
	username, err := getFormValue(r, "username")
	if err != nil {
		return "", "", "", errors.New("Could not parse username.")
	}
	email, err := getFormValue(r, "email")
	if err != nil {
		return "", "", "", errors.New("Could not parse email.")
	}
	password, err := getFormValue(r, "password")
	if err != nil {
		return "", "", "", errors.New("Could not parse password.")
	}
	password2, err := getFormValue(r, "password2")
	if err != nil {
		return "", "", "", errors.New("Could not parse repeated password.")
	}
	if !isLegitUsername(username) {
		return "", "", "", errors.New("Username contains unallowed characters.")
	}
	if !isLegitEmailAddress(email) && email != "" {
		return "", "", "", errors.New("Email address looks illegit.")
	}
	if password != password2 {
		return "", "", "", errors.New("The two passwords to not match.")
	}
	if _, err = getRunnerByUsername(username); err == nil {
		return "", "", "", errors.New("A user with that username already exists.")
	}
	return username, email, password, nil
}

// reportFormParser parses the form for reporting runs, and returns
// the explanation given by the reporter on success
func reportFormParser(r *http.Request) (string, error) {
	err := r.ParseForm()
	if err != nil {
		return "", errors.New("Could not parse form contents.")
	}
	explanation, err := getFormValue(r, "explanation")
	if err != nil || explanation == "" {
		return "", errors.New("Explanation given can not be empty.")
	}
	return explanation, nil
}
