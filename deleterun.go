package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

// deleteRunHandler handles POST requests to "/delete-run". Responds
// with a simple success = true/false as JSON.
func deleteRunHandler(w http.ResponseWriter, r *http.Request) {
	err := parseDeleteRunRequest(r)
	var response []byte
	if err != nil {
		errorMessage := make(map[string]string)
		errorMessage["error"] = err.Error()
		response, _ = json.Marshal(&errorMessage)
	} else {
		successMessage := make(map[string]bool)
		successMessage["success"] = true
		response, _ = json.Marshal(&successMessage)
	}
	w.Write(response)
	return
}

// parseDeleteRunRequest parses requests to "/delete-run"
func parseDeleteRunRequest(r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return errors.New("Could not parse request")
	}
	runIDstr, err := getFormValue(r, "runID")
	if err != nil {
		return errors.New("Could not parse run ID")
	}
	runID, err := strconv.Atoi(runIDstr)
	if err != nil {
		return errors.New("Could not parse run ID")
	}
	run, err := getRunByID(runID)
	if err != nil {
		return errors.New("Could not find run.")
	}
	activeUser, err := getActiveUser(r)
	if err != nil {
		return errors.New("User not logged in")
	}
	if activeUser.ID != run.Runner.ID {
		return errors.New("User is not the runner of the run")
	}
	err = run.delete()
	if err != nil {
		return errors.New("Could not delete run")
	}
	return nil
}
