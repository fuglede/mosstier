package main

import (
	"fmt"
)

type run struct {
	ID             int
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

// getAllWorldRecrods returns a slice of all current world records
func getAllWorldRecords() ([]run, error) {
	var err error
	var runsInCategory []run
	var records []run

	for _, cat := range getAllCategories() {
		runsInCategory, err = getRunsByCategory(cat, 1)
		if err != nil {
			return nil, err
		}
		records = append(records, runsInCategory[0])
	}
	return records, nil
}

// getRunsByCategory returns the top `limit` runs in a given category. If `limit` is 0, returns all runs.
func getRunsByCategory(category category, limit int64) (runs []run, err error) {
	query := "SELECT runs.id, runs.score, runs.level, runs.link, runs.spelunker, runs.date, runs.comment, users.id, users.username, users.country FROM runs INNER JOIN users ON runs.runner = users.id WHERE runs.cat = ? ORDER BY runs.score"
	if category.Goal == "Score" {
		query += " DESC"
	}
	if limit != 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	statement, err := db.Prepare(query)
	if err != nil {
		return
	}
	defer statement.Close()
	rows, err := statement.Query(category.ID)
	if err != nil {
		return
	}
	i := 1
	for rows.Next() {
		var r run
		var p runner
		var spelunkerID int
		err = rows.Scan(&r.ID, &r.Score, &r.Level, &r.Link, &spelunkerID, &r.Time, &r.Comment, &p.ID, &p.Username, &p.Country)
		if err != nil {
			return
		}
		r.Runner = p
		r.Category = category
		r.Spelunker, _ = getSpelunkerByID(spelunkerID)
		r.RankInCategory = i
		runs = append(runs, r)
		i++
	}
	return
}

func getRunsByRunnerID(runnerID int) (runs []run, err error) {
	query := "SELECT id, cat, score, level, link, spelunker, date, comment FROM runs WHERE runner = ? ORDER BY cat"
	statement, err := db.Prepare(query)
	if err != nil {
		return
	}
	defer statement.Close()
	rows, err := statement.Query(runnerID)
	if err != nil {
		return
	}
	for rows.Next() {
		var r run
		var spelunkerID int
		var categoryID int
		err = rows.Scan(&r.ID, &categoryID, &r.Score, &r.Level, &r.Link, &spelunkerID, &r.Time, &r.Comment)
		if err != nil {
			return
		}
		r.Category, _ = getCategoryByID(categoryID)
		r.Spelunker, _ = getSpelunkerByID(spelunkerID)
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
	}
	minutes := r.Score / 60000
	seconds := (r.Score - 60000*minutes) / 1000
	millisecs := r.Score - 60000*minutes - 1000*seconds
	return fmt.Sprintf("%d:%02d:%03d", minutes, seconds, millisecs)
}
