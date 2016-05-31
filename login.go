package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
)

// Login sessions are handled using a single cookie store. Keys
// are updated on a per-server launch basis.
var cookieStore *sessions.CookieStore

func initializeCookieStore() {
	cookieStore = sessions.NewCookieStore([]byte("hest"))

	//cookieStore = sessions.NewCookieStore(securecookie.GenerateRandomKey(64),
	//	securecookie.GenerateRandomKey(32))
	cookieStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   0, // Lasts until end of session
		HttpOnly: true,
	}
}

// getActiveUser returns the currently logged in user, if any.
func getActiveUser(r *http.Request) (user runner, err error) {
	session, err := cookieStore.Get(r, "login")
	if err != nil {
		return
	}
	storedRunnerID, ok := session.Values["userID"]
	if !ok {
		err = errors.New("no user ID found in cookie")
		return
	}
	// Session values are map[string]interface{}, so we need to
	// cast and check the type.
	runnerID, ok := storedRunnerID.(int)
	if !ok {
		err = errors.New("user ID was not an integer")
		return
	}
	user, err = getRunnerByID(runnerID)
	return
}

// setActiveUser sets the currently logged in user.
func setActiveUser(r *http.Request, w http.ResponseWriter, user runner) (err error) {
	session, err := cookieStore.Get(r, "login")
	if err != nil {
		return
	}
	session.Values["userID"] = user.ID
	session.Save(r, w)
	return
}

// generatePassword generates a 25 byte long random password.
func generatePassword() (string, error) {
	b := make([]byte, 25)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// The slice should now contain random bytes instead of only zeroes.
	// Base64 it just to have something users can use.
	return base64.StdEncoding.EncodeToString(b), err
}
