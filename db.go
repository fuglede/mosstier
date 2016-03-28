package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var db *sql.DB

func initializeDatabase() {
	db, err := sql.Open("mysql", config.DbConnection)
	if err != nil {
		log.Fatal("Could not open database.")
	}
	defer db.Close()
}