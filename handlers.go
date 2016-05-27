package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// aboutHandler handles GET requests to "/about"
func aboutHandler(w http.ResponseWriter, r *http.Request) {
	renderContent("tmpl/about.html", w, nil)
}

// categoryHandler handles GET requests to "/category*"
func categoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cat, err := getCategoryByAbbr(vars["categoryName"])
	if err != nil {
		http.NotFound(w, r)
		return
	}
	runs, err := getRunsByCategory(cat, 0)
	if err != nil {
		log.Println("Could not get runs: ", err)
		http.Error(w, "Internal server error", 500)
	}
	highlightedRunner := vars["runner"]
	type categoryData struct {
		Category          category
		Runs              []run
		HighlightedRunner string
	}
	data := categoryData{cat, runs, highlightedRunner}
	renderContent("tmpl/category.html", w, data)
}

// contactHandler handles GET and POST requests to "/contact"
func contactHandler(w http.ResponseWriter, r *http.Request) {
	// We will record any form errors in a single string.
	type contactData struct {
		MailSent bool
		Error    string
	}
	var mailSent = false
	var errorString string
	if r.Method == "POST" {
		success := true
		err := r.ParseForm()
		if err != nil {
			errorString += "Could not read form contents. "
			success = false
		}
		if len(r.Form["name"][0]) == 0 {
			errorString += "Name field can not be empty. "
			success = false
		}
		if len(r.Form["subject"][0]) == 0 {
			errorString += "Subject field can not be empty. "
			success = false
		}
		if len(r.Form["message"][0]) == 0 {
			errorString += "Message field can not be empty. "
			success = false
		}
		if success {
			subject := "Moss Tier contact form message: " + r.Form["subject"][0]
			message := "From: " + r.Form["name"][0] + "\r\n"
			if r.Form["email"] != nil {
				message += "Email: " + r.Form["email"][0] + "\r\n"
			}
			message += "\r\n\r\n" + r.Form["message"][0]
			err = sendMail(config.AdminEmail, subject, message)
			if err != nil {
				errorString = "Mail delivery failed."
				success = false
			} else {
				mailSent = true
			}
		}
	}
	renderContent("tmpl/contact.html", w, contactData{mailSent, errorString})
}

// frontPageHandler handles GET requests to "/"
func frontPageHandler(w http.ResponseWriter, r *http.Request) {
	type frontPageData struct {
		News    []newsEntry
		Records []run
	}
	worldRecords, err := getAllWorldRecords()
	if err != nil {
		log.Println("Could not get world records: ", err)
		http.Error(w, "Internal server error", 500)
		return
	}
	data := frontPageData{readNews(), worldRecords}
	renderContent("tmpl/frontpage.html", w, data)
}

// passwordResetHandler handles GET requests to "/password-reset"
func passwordResetHandler(w http.ResponseWriter, r *http.Request) {
	type passwordResetData struct {
		PasswordReset bool
		Error         string
	}
	passwordReset := false
	var errorString string

	if r.Method == "POST" {
		formValid := true
		err := r.ParseForm()
		if err != nil {
			errorString += "Could not parse form contents. "
			formValid = false
		}
		if len(r.Form["username"][0]) == 0 {
			errorString += "Username entry can not be empty. "
			formValid = false
		}
		if len(r.Form["email"][0]) == 0 {
			errorString += "Email entry can not be empty. "
			formValid = false
		}
		if formValid {
			user, err := getRunnerByUsernameAndEmail(r.Form["username"][0], r.Form["email"][0])
			if err != nil {
				errorString += "Could not find any user with that combination of username and email. "
			} else {
				newPassword := generatePassword()
				// Send a mail to the user with the password, before attempting to update it
				err = user.sendMail("Password reset", "Hi "+user.Username+". Someone (hopefully you) requested "+
					"a new password for you on Moss Tier. Here's your new one: "+newPassword)
				if err != nil {
					errorString += "Could not send you an email with your new password."
					log.Println(err)
				} else {
					err = user.updatePassword(generatePassword())
					if err != nil {
						errorString += "Could not update your password due to an unexpected error. "
						log.Println(err)
					} else {
						passwordReset = true
					}
				}
			}
		}
	}
	renderContent("tmpl/passwordreset.html", w, passwordResetData{passwordReset, errorString})
}

// profileHandler handles GET requests to "/profile*"
func profileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	profileID, err := strconv.Atoi(vars["profileID"])
	if err != nil {
		log.Println("Could not parse profile id: ", err)
		http.NotFound(w, r)
		return
	}
	thisRunner, err := getRunnerByID(profileID)
	if err != nil {
		log.Println("Could not find runner: ", err)
		http.NotFound(w, r)
		return
	}
	runs, err := getRunsByRunnerID(profileID)
	if err != nil {
		log.Println("Could not get runs: ", err)
		http.Error(w, "Internal server error", 500)
	}
	type profileData struct {
		Runner runner
		Runs   []run
	}
	data := profileData{thisRunner, runs}
	renderContent("tmpl/profile.html", w, data)
}

// registerHandler handles GET and POST requests to "/register"
func registerHandler(w http.ResponseWriter, r *http.Request) {
	type registerData struct {
		Success bool
		Error   string
	}
	success := false
	errorString := ""
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			errorString += "Could not parse form contents. "
		} else {
			username := r.Form["username"][0]
			email := r.Form["email"][0]
			password := r.Form["password"][0]
			password2 := r.Form["password2"][0]
			shouldCreateUser := true
			if !isLegitUsername(username) {
				errorString += "Username contains unallowed characters. "
				shouldCreateUser = false
			}
			if !isLegitEmailAddress(email) && email != "" {
				errorString += "Email address looks illegit. "
				shouldCreateUser = false
			}
			if password != password2 {
				errorString += "The two passwords to not match. "
				shouldCreateUser = false
			}
			if _, err = getRunnerByUsername(username); err == nil {
				errorString += "A user with that username already exists. "
				shouldCreateUser = false
			}
			if shouldCreateUser {
				err = makeUser(username, email, password)
				if err != nil {
					errorString += "Could not create user. Please try again later. "
					log.Println(err)
				} else {
					success = true
				}
			}
		}
	}

	data := registerData{success, errorString}
	renderContent("tmpl/register.html", w, data)
}

// rulesHandler handles GET requests to "/rules"
func rulesHandler(w http.ResponseWriter, r *http.Request) {
	renderContent("tmpl/rules.html", w, getAllCategories())
}
