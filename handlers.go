package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// aboutHandler handles GET requests to "/about"
func aboutHandler(w http.ResponseWriter, r *http.Request) {
	renderContent("tmpl/about.html", r, w, nil)
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
		return
	}
	highlightedRunner := vars["runner"]
	type categoryData struct {
		Category          category
		Runs              []run
		HighlightedRunner string
	}
	data := categoryData{cat, runs, highlightedRunner}
	renderContent("tmpl/category.html", r, w, data)
}

// contactHandler handles GET and POST requests to "/contact"
func contactHandler(w http.ResponseWriter, r *http.Request) {
	// We keep track of user input in the template data to provide
	// a partially filled form in case the user messes up the inputs
	type contactData struct {
		MailSent     bool
		Error        string
		NameInput    string
		EmailInput   string
		SubjectInput string
		MessageInput string
	}
	var mailSent = false
	var errorString string
	var name string
	var email string
	var subject string
	var message string
	var err error
	if r.Method == "POST" {
		name, email, subject, message, err = contactFormParser(r)
		if err != nil {
			errorString = err.Error()
		} else {
			mailSubject := "Moss Tier contact form message: " + subject
			mailBody := "From: " + name + "\r\n"
			if email != "" {
				mailBody += "Email: " + email + "\r\n"
			}
			mailBody += "\r\n\r\n" + message
			err = sendMail(config.AdminEmail, mailSubject, mailBody)
			if err != nil {
				errorString = "Mail delivery failed."
			} else {
				mailSent = true
			}
		}
	}
	data := contactData{mailSent, errorString, name, email, subject, message}
	renderContent("tmpl/contact.html", r, w, data)
}

// editProfileHandler handles GET requests to "/edit-profile"
func editProfileHandler(w http.ResponseWriter, r *http.Request) {
	// First, let's make sure that the user is logged in
	if _, err := getActiveUser(r); err != nil {
		http.NotFound(w, r)
		return
	}
	type editProfileData struct {
		Success    bool
		Error      string
		Countries  map[string]string
		Spelunkers []spelunker
	}
	success := false
	var errorString string
	data := editProfileData{success, errorString, countries, spelunkers}
	renderContent("tmpl/editprofile.html", r, w, data)
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
	renderContent("tmpl/frontpage.html", r, w, data)
}

// loginHandler handles GET and POST requests to "/login"
func loginHandler(w http.ResponseWriter, r *http.Request) {
	type loginData struct {
		Success       bool
		Error         string
		UsernameInput string
		PasswordInput string
	}
	success := false
	var errorString string
	var username string
	var password string
	var user runner
	var err error

	if r.Method == "POST" {
		username, password, user, err = loginFormParser(r)
		if err != nil {
			errorString = err.Error()
		} else {
			err = setActiveUser(r, w, user)
			if err != nil {
				log.Println(err)
				http.Error(w, "Internal server error", 500)
				return
			}
			success = true
		}
	}

	data := loginData{success, errorString, username, password}

	renderContent("tmpl/login.html", r, w, data)
}

// logOutHandler handles GET requests to "/log-out"
func logOutHandler(w http.ResponseWriter, r *http.Request) {
	// A natural way to log out would be to set the session
	// cookie max-age to -1, but that adds an unnecessary layer
	// of complexity for when the user wants to log back in.
	// Instead we just change the user to one that doesn't exist.
	session, err := cookieStore.Get(r, "login")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", 500)
		return
	}
	session.Values["userID"] = 0
	session.Save(r, w)
	renderContent("tmpl/logout.html", r, w, nil)
}

// passwordResetHandler handles GET and POST requests to "/password-reset"
func passwordResetHandler(w http.ResponseWriter, r *http.Request) {
	type passwordResetData struct {
		PasswordReset bool
		Error         string
	}
	passwordReset := false
	var errorString string

	if r.Method == "POST" {
		user, err := passwordResetFormParser(r)
		if err != nil {
			errorString = err.Error()
		} else {
			newPassword, err := generatePassword()
			if err != nil {
				log.Println(err)
				http.Error(w, "Internal server error", 500)
				return
			}
			// Send a mail to the user with the password, before attempting to update it
			err = user.sendMail("Password reset", "Hi "+user.Username+". Someone (hopefully you) requested "+
				"a new password for you on Moss Tier. Here's your new one: "+newPassword)
			if err != nil {
				errorString += "Could not send you an email with your new password."
				log.Println(err)
			} else {
				err = user.updatePassword(newPassword)
				if err != nil {
					errorString += "Could not update your password due to an unexpected error. "
					log.Println(err)
				} else {
					passwordReset = true
				}
			}
		}
	}
	renderContent("tmpl/passwordreset.html", r, w, passwordResetData{passwordReset, errorString})
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
		Runner *runner
		Runs   []run
	}
	data := profileData{&thisRunner, runs}
	renderContent("tmpl/profile.html", r, w, data)
}

// registerHandler handles GET and POST requests to "/register"
func registerHandler(w http.ResponseWriter, r *http.Request) {
	type registerData struct {
		Success        bool
		Error          string
		UsernameInput  string
		EmailInput     string
		PasswordInput  string
		Password2Input string
	}
	success := false
	var errorString string
	var username string
	var email string
	var password string
	var password2 string
	var err error
	if r.Method == "POST" {
		username, email, password, password2, err = registerFormParser(r)
		if err != nil {
			errorString = err.Error()
		} else {
			err = makeUser(username, email, password)
			if err != nil {
				errorString += "Could not create user. Please try again later. "
				log.Println(err)
			} else {
				// Everything is good; now, log in as the newly
				// created user
				user, err := getRunnerByUsername(username)
				if err != nil {
					log.Println(err)
					http.Error(w, "Internal server error", 500)
					return
				}
				err = setActiveUser(r, w, user)
				if err != nil {
					log.Println(err)
					http.Error(w, "Internal server error", 500)
					return
				}
				success = true
			}
		}
	}

	data := registerData{success, errorString, username,
		email, password, password2}
	renderContent("tmpl/register.html", r, w, data)
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
		explanation, err := reportFormParser(r)
		if err != nil {
			errorString = err.Error()
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

	data := reportData{run, success, errorString}
	renderContent("tmpl/report.html", r, w, data)
}

// rulesHandler handles GET requests to "/rules"
func rulesHandler(w http.ResponseWriter, r *http.Request) {
	renderContent("tmpl/rules.html", r, w, getAllCategories())
}
