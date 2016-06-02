package main

import (
	"errors"
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
	query := "SELECT runs.id, runs.score, runs.level, runs.link, runs.spelunker, runs.date, runs.comment, users.id, users.username, users.country FROM runs INNER JOIN users ON runs.runner = users.id WHERE runs.cat = ? AND runs.flag = '' ORDER BY runs.score"
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

// getRunsByRunnerID produces a slice of all runs registered for a given runner,
func getRunsByRunnerID(runnerID int) (runs []run, err error) {
	runner, err := getRunnerByID(runnerID)
	if err != nil {
		return
	}
	query := "SELECT id, cat, score, level, link, spelunker, date, comment, flag FROM runs WHERE runner = ? ORDER BY cat"
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
		err = rows.Scan(&r.ID, &categoryID, &r.Score, &r.Level, &r.Link, &spelunkerID, &r.Time, &r.Comment, &r.Flag)
		if err != nil {
			return
		}
		r.Runner = runner
		r.Category, _ = getCategoryByID(categoryID)
		r.Spelunker, _ = getSpelunkerByID(spelunkerID)
		runs = append(runs, r)
	}
	return
}

// getRunByID returns the run with a given integral ID.
func getRunByID(runID int) (r run, err error) {
	stmt, err := db.Prepare("SELECT runs.score, runs.cat, runs.level, runs.link, runs.spelunker, runs.comment, users.username FROM runs INNER JOIN users ON runs.runner = users.id WHERE runs.id = ?")
	if err != nil {
		return
	}
	defer stmt.Close()
	var categoryID int
	var spelunkerID int
	err = stmt.QueryRow(runID).Scan(&r.Score, &categoryID, &r.Level, &r.Link, &spelunkerID, &r.Comment, &r.Runner.Username)
	r.Runner, _ = getRunnerByUsername(r.Runner.Username)
	r.ID = runID
	r.Category, _ = getCategoryByID(categoryID)
	r.Spelunker, _ = getSpelunkerByID(spelunkerID)
	return
}

// Flag flags the run, removing it from the leaderboards, and
// informing the runner the reason why.
func (r *run) flag(reason string) error {
	query, err := db.Prepare("UPDATE runs SET flag = ? WHERE id = ?")
	if err != nil {
		return errors.New("Could not prepare database: " + err.Error())
	}
	_, err = query.Exec(reason, r.ID)
	if err != nil {
		return errors.New("Could not perform database query: " + err.Error())
	}
	// Now inform the user if they have asked to be informed
	fmt.Println(r.Runner.EmailFlag)
	fmt.Println(r.Runner)
	if r.Runner.EmailFlag {
		mailBody := "Hi %s.\n\nThis is to inform you that your Moss Tier run " +
			"in the category %s has been flagged as violating the rules by one " +
			"of the moderators. The reason they gave was the following:\n\n%s"
		err = r.Runner.sendMail("Moss Tier run flagged",
			fmt.Sprintf(mailBody, r.Runner.Username, r.Category.Name, reason))
		if err != nil {
			return errors.New("Flagged run but could not inform user: " + err.Error())
		}
	}
	return nil
}

// GetWorld returns the last world, the player was in during the run
// as an integer between 1 and 5.
func (r *run) GetWorld() int {
	return (r.Level-1)/4 + 1
}

// GetWorld returns the last subworld floor, the player was in during
// the run as an integer between 1 and 4.
func (r *run) GetFloor() int {
	return (r.Level-1)%4 + 1
}

// FormatLevel takes the (one-indexed) number of a level (e.g. 5) and produces
// a string describing it (e.g. 2-1).
func (r *run) FormatLevel() string {
	return fmt.Sprintf("%d-%d", r.GetWorld(), r.GetFloor())
}

// NumberOfMinutes returns the number of minutes spent in the run (rounded down),
// assuming implicitly that the run is a speed run.
func (r *run) NumberOfMinutes() int {
	return r.Score / 60000
}

// NumberOfSeconds returns the number of seconds spent in the run (rounded down),
// within the last minute of the run, assuming implicitly that the run is a
// speed run.
func (r *run) NumberOfSeconds() int {
	return (r.Score - 60000*r.NumberOfMinutes()) / 1000
}

// NumberOfMilliseconds returns the number of milliseconds spent in the run,
// within the last second of the run, assuming implicitly that the run is a
// speed run.
func (r *run) NumberOfMilliseconds() int {
	return r.Score - 60000*r.NumberOfMinutes() - 1000*r.NumberOfSeconds()
}

// FormatScore turns a result type integer into a readable result, either by adding
// a dollar sign to a score, or by turning a number of milliseconds into a formatted time.
func (r *run) FormatScore() string {
	if r.Category.Goal == "Score" {
		return fmt.Sprintf("$%d", r.Score)
	}
	return fmt.Sprintf("%d:%02d:%03d",
		r.NumberOfMinutes(),
		r.NumberOfSeconds(),
		r.NumberOfMilliseconds())
}
