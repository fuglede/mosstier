package main

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
)

const scoreLeaderboards = "https://steamcommunity.com/stats/239350/leaderboards/164848/?xml=1"
const speedLeaderboards = "https://steamcommunity.com/stats/239350/leaderboards/164849/?xml=1"

type steamResponse struct {
	XMLName xml.Name `xml:"response"`
	Entries []steamRunEntry  `xml:"entries>entry"`
}

type steamRunEntry struct {
	SteamId string `xml:"steamid"`
	Score   string `xml:"score"`
	Details string `xml:"details"`
}

// getResultFromSteamLeaderboards produces from a Steam user ID the best result
// the given user has obtained in a run of a given type, if that user is in the
// top 5000. It also returns the spelunker used for that run, as well as the final
// level the user was in in that run.
func getResultFromSteamLeaderboards(steamId int64, runType string) (result int64, level int64, spelunker int64, err error) {
	result = -1
	level = -1
	spelunker = -1
	
	var leaderboardsUrl string
	if runType == "score" {
		leaderboardsUrl = scoreLeaderboards
	} else if runType == "speed" {
		leaderboardsUrl = speedLeaderboards
	}

	response, err := http.Get(leaderboardsUrl)
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
		if userId, _ := strconv.ParseInt(entry.SteamId, 10, 64); userId == steamId {
			result, _ = strconv.ParseInt(entry.Score, 10, 64)
			// The Steam "details" response is a string containing (hexadecimally represented)
			// substrings describing the ending level and spelunker used in a given run.
			level, _ = strconv.ParseInt(entry.Details[:2], 16, 64)
			spelunker, _ = strconv.ParseInt(entry.Details[8:10], 16, 64)
			return
		}
	}
	err = errors.New("Steam ID not found in top 5000")
	return
}