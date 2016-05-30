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
func contactFormParser(r *http.Request) (name string, email string,
	subject string, message string, err error) {
	err = r.ParseForm()
	if err != nil {
		return "", "", "", "", errors.New("Could not read form contents.")
	}
	name, nameErr := getFormValue(r, "name")
	email, emailErr := getFormValue(r, "email")
	subject, subjectErr := getFormValue(r, "subject")
	message, messageErr := getFormValue(r, "message")
	if nameErr != nil || name == "" {
		err = errors.New("Name field can not be empty.")
		log.Println(subject)
		return
	}
	if emailErr != nil {
		err = errors.New("Could not parse email.")
		return
	}
	if subjectErr != nil || subject == "" {
		err = errors.New("Subject field can not be empty.")
		return
	}
	if messageErr != nil || message == "" {
		err = errors.New("Message field can not be empty.")
		return
	}
	return
}

// loginFormParser parses POST requests to "/login". Returns the
// user to log in on success, and the form contents in either case.
func loginFormParser(r *http.Request) (username string, password string,
	user runner, err error) {
	err = r.ParseForm()
	if err != nil {
		err = errors.New("Could not parse form contents.")
		return
	}
	username, usernameErr := getFormValue(r, "username")
	password, passwordErr := getFormValue(r, "password")
	if usernameErr != nil || !isLegitUsername(username) {
		err = errors.New("Invalid username.")
		return
	}
	if passwordErr != nil || !isLegitPassword(password) {
		err = errors.New("Invalid password.")
		return
	}
	user, err = getRunnerByUsername(username)
	if err != nil {
		err = errors.New("Could not find a user with that username.")
		return
	}
	err = user.testLogin(password)
	if err != nil {
		err = errors.New("Incorrect password.")
	}
	return
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
// the posted contents.
func registerFormParser(r *http.Request) (username string, email string,
	password string, password2 string, err error) {
	err = r.ParseForm()
	if err != nil {
		err = errors.New("Could not parse form contents.")
		return
	}
	username, usernameErr := getFormValue(r, "username")
	email, emailErr := getFormValue(r, "email")
	password, passwordErr := getFormValue(r, "password")
	password2, password2Err := getFormValue(r, "password2")
	if usernameErr != nil {
		err = errors.New("Could not parse username.")
		return
	}
	if emailErr != nil {
		err = errors.New("Could not parse email.")
		return
	}
	if passwordErr != nil {
		err = errors.New("Could not parse password.")
		return
	}
	if password2Err != nil {
		err = errors.New("Could not parse repeated password.")
		return
	}
	if !isLegitUsername(username) {
		err = errors.New("Username contains unallowed characters.")
		return
	}
	if !isLegitEmailAddress(email) && email != "" {
		err = errors.New("Email address looks illegit.")
		return
	}
	if password != password2 {
		err = errors.New("The two passwords to not match.")
		return
	}
	if _, err = getRunnerByUsername(username); err == nil {
		err = errors.New("A user with that username already exists.")
		return
	}
	err = nil
	return
}

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
