package main

import (
	"fmt"
)

type run struct {
	Id int
	Runner int
	Category int
	Score int
	Level int
	Link string
	Platform int
	Spelunker int
	Date int
	Comment string
	Flag string
}

// getRunsByCategory returns all runs in a given category
func getRunsByCategory(category category) (runs []run, err error) {
	fmt.Println("3")
	fmt.Println("%v", db)
	statement, err := db.Prepare("SELECT * FROM runs WHERE cat = ?")
	fmt.Println("4")
	if err != nil {
		return
	}
	defer statement.Close()
	rows, err := statement.Query(category.Id)
	fmt.Println("2")
	if err != nil {
		return
	}
	for rows.Next() {
		var r run
		err = rows.Scan(&r.Id, &r.Runner, &r.Category, &r.Score, &r.Level, &r.Link, &r.Platform, &r.Spelunker, &r.Date, &r.Comment, &r.Flag)
		runs = append(runs, r)
		if err != nil {
			return
		}
	}
	return
}