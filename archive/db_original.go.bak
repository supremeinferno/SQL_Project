package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func connectDB() {
	var err error
	db, err = sql.Open("sqlite3", "./hostel.db")
	if err != nil {
		panic(err)
	}
}
