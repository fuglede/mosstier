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
		err := r.ParseForm()
		if err != nil {
			errorString += "Could not read form contents. "
		}
		name, err := getFormValue(r, "name")
		if err != nil {
			errorString += "Could not parse name. "
		}
		if len(name) == 0 {
			errorString += "Name field can not be empty. "
		}
		// We try to parse the email, even if it is not required;
		// if not given, it should still return an empty string.
		email, err := getFormValue(r, "email")
		if err != nil {
			errorString += "Could not parse email. "
		}
		subject, err := getFormValue(r, "subject")
		if err != nil {
			errorString += "Could not parse subject. "
		}
		if len(subject) == 0 {
			errorString += "Subject field can not be empty. "
		}
		message, err := getFormValue(r, "message")
		if err != nil {
			errorString += "Could not parse message. "
		}
		if len(message) == 0 {
			errorString += "Message field can not be empty. "
		}
		if errorString == "" {
			subject := "Moss Tier contact form message: " + subject
			mailBody := "From: " + name + "\r\n"
			if email != "" {
				mailBody += "Email: " + email + "\r\n"
			}
			mailBody += "\r\n\r\n" + message
			err = sendMail(config.AdminEmail, subject, mailBody)
			if err != nil {
				errorString = "Mail delivery failed."
			} else {
				mailSent = true
			}
		}
	}
	renderContent("tmpl/contact.html", w, contactData{mailSent, errorString})
}

// frontPageHandler handles GET requests to "/"
func frontPageHandler(w http.ResponseWriter, r *http.Request) {
	// We split the world records into t heir two classes. Rather
	// than just using map to do this, we make our own struct to
	// be able to control ordering more explicitly.
	type classWithRecords struct {
		Description string
		Records     []run
	}
	type frontPageData struct {
		News         []newsEntry
		WorldRecords []classWithRecords
	}
	allWorldRecords, err := getAllWorldRecords()
	var mainWRs []run
	var challengeWRs []run
	if err != nil {
		log.Println("Could not get world records: ", err)
		http.Error(w, "Internal server error", 500)
		return
	}
	// Now actually split split the world records
	for _, wr := range allWorldRecords {
		if wr.Category.isMain() {
			mainWRs = append(mainWRs, wr)
		} else {
			challengeWRs = append(challengeWRs, wr)
		}
	}

	worldRecords := []classWithRecords{
		classWithRecords{"Main categories", mainWRs},
		classWithRecords{"Challenge categories", challengeWRs},
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
		err := r.ParseForm()
		if err != nil {
			errorString += "Could not parse form contents. "
		}
		username, err := getFormValue(r, "username")
		if err != nil {
			errorString += "Could not parse username. "
		}
		if username == "" {
			errorString += "Username entry can not be empty. "
		}
		email, err := getFormValue(r, "email")
		if err != nil {
			errorString += "Could not parse email. "
		}
		if email == "" {
			errorString += "Email entry can not be empty. "
		}
		if errorString == "" {
			user, err := getRunnerByUsernameAndEmail(username, email)
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
			username, err := getFormValue(r, "username")
			if err != nil {
				errorString += "Could not parse username. "
			}
			email, err := getFormValue(r, "email")
			if err != nil {
				errorString += "Could not parse email. "
			}
			password, err := getFormValue(r, "password")
			if err != nil {
				errorString += "Could not parse password. "
			}
			password2, err := getFormValue(r, "password2")
			if err != nil {
				errorString += "Could not parse repeated password. "
			}
			if !isLegitUsername(username) {
				errorString += "Username contains unallowed characters. "
			}
			if !isLegitEmailAddress(email) && email != "" {
				errorString += "Email address looks illegit. "
			}
			if password != password2 {
				errorString += "The two passwords to not match. "
			}
			if _, err = getRunnerByUsername(username); err == nil {
				errorString += "A user with that username already exists. "
			}
			if errorString == "" {
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

// reportHandler handles GET and POST requests to /report*
func reportHandler(w http.ResponseWriter, r *http.Request) {
	type reportData struct {
		Run     run
		Success bool
		Error   string
	}
	success := false
	var errorString string

	vars := mux.Vars(r)
	runID, err := strconv.Atoi(vars["runID"])
	if err != nil {
		log.Println("Could not parse run ID: ", err)
		http.NotFound(w, r)
		return
	}
	run, err := getRunByID(runID)
	if err != nil {
		log.Println("Could not find run with given ID: ", err)
		http.NotFound(w, r)
		return
	}

	if r.Method == "POST" {
		err = r.ParseForm()
		if err != nil {
			errorString += "Could not parse form contents. "
		} else {
			explanation, err := getFormValue(r, "explanation")
			if err != nil {
				log.Println("Invalid explanation: ", err)
				http.Error(w, "Internal server error", 500)
				return
			}
			if explanation == "" {
				errorString += "Explanation given can not be empty. "
			} else {
				err = sendMails(config.Moderators, "Moss Tier run reported",
					"Hi Moss Tier moderator. The run by "+run.Runner.Username+" in the "+
						"category "+run.Category.Name+" (id "+strconv.Itoa(runID)+") "+
						"has been reported as violating the rules. Could you check it "+
						"out and flag the run if needed? The explanation they gave was "+
						"\""+explanation+"\"")
				if err != nil {
					errorString += "Could not send mail to moderators. Please try again later. "
				} else {
					success = true
				}
			}
		}
	}

	data := reportData{run, success, errorString}
	renderContent("tmpl/report.html", w, data)
}

// rulesHandler handles GET requests to "/rules"
func rulesHandler(w http.ResponseWriter, r *http.Request) {
	renderContent("tmpl/rules.html", w, getAllCategories())
}
