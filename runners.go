package main

import ()

type runner struct {
	Id             int
	Username       string
	Email          string
	Country        string
	Spelunker      int
	Steam          int
	Psn            string
	Xbla           string
	Twitch         string
	YouTube        string
	FreeText       string
	EmailFlag      int
	EmailWr        int
	EmailChallenge int
}

func getRunnerById(id int) (r runner, err error) {
	query := "SELECT id, username, country, spelunker, steam, psn, xbla, twitch, youtube, freetext FROM users WHERE id = ?"
	statement, err := db.Prepare(query)
	if err != nil {
		return
	}
	defer statement.Close()
	err = statement.QueryRow(id).Scan(&r.Id, &r.Username, &r.Country, &r.Spelunker, &r.Steam, &r.Psn, &r.Xbla, &r.Twitch, &r.YouTube, &r.FreeText)
	return
}
