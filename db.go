package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func initializeDatabase() {
	db, _ = sql.Open("mysql", config.DbConnection)
	defer db.Close()
}