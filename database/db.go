package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func ConnectDB() {
	var err error

	// Open SQLite DB
	DB, err = sql.Open("sqlite3", "./hostel.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	// Enable foreign key constraints
	_, err = DB.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Fatal("Failed to enable foreign keys:", err)
	}

	// Test DB connection
	err = DB.Ping()
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	log.Println("✅ Database connected (SQLite)")
}
