package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
)

const scoreLeaderboards = "https://steamcommunity.com/stats/239350/leaderboards/164848/?xml=1"
const speedLeaderboards = "https://steamcommunity.com/stats/239350/leaderboards/164849/?xml=1"

type steamResponse struct {
	XMLName xml.Name        `xml:"response"`
	Entries []steamRunEntry `xml:"entries>entry"`
}

type steamRunEntry struct {
	SteamID string `xml:"steamid"`
	Score   string `xml:"score"`
	Details string `xml:"details"`
}

// getResultFromSteamLeaderboards produces from a Steam user ID the best result
// the given user has obtained in a run of a given type, if that user is in the
// top 5000. It also returns the spelunker used for that run, as well as the final
// level the user was in in that run.
func getResultFromSteamLeaderboards(steamID int, runType string) (result int, level int, spelunker int, err error) {
	result = -1
	level = -1
	spelunker = -1

	var leaderboardsURL string
	if runType == "score" {
		leaderboardsURL = scoreLeaderboards
	} else if runType == "speed" {
		leaderboardsURL = speedLeaderboards
	}

	response, err := http.Get(leaderboardsURL)
	if err != nil {
		return
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	xmlFromSteam := steamResponse{}
	err = xml.Unmarshal(contents, &xmlFromSteam)
	if err != nil {
		return
	}
	for _, entry := range xmlFromSteam.Entries {
		if userID, _ := strconv.Atoi(entry.SteamID); userID == steamID {
			result, _ = strconv.Atoi(entry.Score)
			// The Steam "details" response is a string containing (hexadecimally represented)
			// substrings describing the ending level and spelunker used in a given run.
			spelunker64, _ := strconv.ParseInt(entry.Details[:2], 16, 64)
			spelunker = int(spelunker64)
			level64, _ := strconv.ParseInt(entry.Details[8:10], 16, 64)
			level = int(level64)
			return
		}
	}
	err = errors.New("Steam ID not found in top 5000")
	return
}

// steamLookupHandler handles POST requests to "/steam-lookup"
func steamLookupHandler(w http.ResponseWriter, r *http.Request) {
	type jsonResponse struct {
		Result      int
		Level       int
		SpelunkerID int
	}
	var response []byte
	result, level, spelunker, err := parseSteamRequest(r)
	if err != nil {
		errorMessage := make(map[string]string)
		errorMessage["error"] = err.Error()
		response, _ = json.Marshal(&errorMessage)
	} else {
		response, _ = json.Marshal(&jsonResponse{result, level, spelunker})
	}
	w.Write(response)
	return
}

// parseSteamRequest performs error handling on all request
// to "/steam-lookup"
func parseSteamRequest(r *http.Request) (result int, level int, spelunker int, err error) {
	err = r.ParseForm()
	if err != nil {
		err = errors.New("Could not parse request.")
		return
	}
	runType, err := getFormValue(r, "runType")
	if err != nil {
		err = errors.New("Could not parse run type.")
		return
	}
	if runType != "score" && runType != "speed" {
		err = errors.New("Unknown run type.")
		return
	}
	user, err := getActiveUser(r)
	if err != nil {
		err = errors.New("User not logged in.")
		return
	}
	if user.Steam == 0 {
		err = errors.New("User has no Steam ID.")
		return
	}
	result, level, spelunker, err = getResultFromSteamLeaderboards(user.Steam, runType)
	return
}
