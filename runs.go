package main

import (
	"fmt"
)

type run struct {
	Id             int
	RankInCategory int
	Runner         runner
	Category       category
	Score          int
	Level          int
	Link           string
	Platform       int
	Spelunker      spelunker
	Time           int
	Comment        string
	Flag           string
}

// getRunsByCategory returns all runs in a given category
func getRunsByCategory(category category) (runs []run, err error) {
	query := "SELECT runs.id, runs.score, runs.level, runs.link, runs.spelunker, runs.date, runs.comment, users.id, users.username, users.country FROM runs INNER JOIN users ON runs.runner = users.id WHERE runs.cat = ? ORDER BY runs.score"
	if category.Goal == "Score" {
		query += " DESC"
	}
	statement, err := db.Prepare(query)
	if err != nil {
		return
	}
	defer statement.Close()
	rows, err := statement.Query(category.Id)
	if err != nil {
		return
	}
	i := 1
	for rows.Next() {
		var r run
		var p runner
		var spelunkerId int
		err = rows.Scan(&r.Id, &r.Score, &r.Level, &r.Link, &spelunkerId, &r.Time, &r.Comment, &p.Id, &p.Username, &p.Country)
		if err != nil {
			return
		}
		r.Runner = p
		r.Category = category
		r.Spelunker, _ = getSpelunkerById(spelunkerId)
		r.RankInCategory = i
		runs = append(runs, r)
		i += 1
	}
	return
}

func getRunsByRunnerId(runnerId int) (runs []run, err error) {
	query := "SELECT id, cat, score, level, link, spelunker, date, comment FROM runs WHERE runner = ? ORDER BY cat"
	statement, err := db.Prepare(query)
	if err != nil {
		return
	}
	defer statement.Close()
	rows, err := statement.Query(runnerId)
	if err != nil {
		return
	}
	for rows.Next() {
		var r run
		var spelunkerId int
		var categoryId int
		err = rows.Scan(&r.Id, &categoryId, &r.Score, &r.Level, &r.Link, &spelunkerId, &r.Time, &r.Comment)
		if err != nil {
			return
		}
		r.Category, _ = getCategoryById(categoryId)
		r.Spelunker, _ = getSpelunkerById(spelunkerId)
		runs = append(runs, r)
	}
	return
}

// FormatLevel takes the (one-indexed) number of a level (e.g. 5) and produces
// a string describing it (e.g. 2-1).
func (r *run) FormatLevel() string {
	world := (r.Level-1)/4 + 1
	floor := (r.Level-1)%4 + 1
	return fmt.Sprintf("%d-%d", world, floor)
}

// FormatScore turns a result type integer into a readable result, either by adding
// a dollar sign to a score, or by turning a number of milliseconds into a formatted time. 
func (r *run) FormatScore() string {
	if r.Category.Goal == "Score" {
		return fmt.Sprintf("$%d", r.Score)
	} else {
		minutes := r.Score / 60000
		seconds := (r.Score - 60000*minutes) / 1000
		millisecs := r.Score - 60000*minutes - 1000*seconds
		return fmt.Sprintf("%d:%02d:%03d", minutes, seconds, millisecs)
	}
}
