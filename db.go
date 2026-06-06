package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/sijms/go-ora/v2"
)

var db *sql.DB

func connectDB() {
	// Read DSN from env (optional)
	dsn := os.Getenv("ORACLE_DSN")

	// Default DSN (Docker Oracle XE)
	if dsn == "" {
		dsn = "oracle://system:Oracle123@localhost:1521/FREEPDB1"
	}

	var err error

	// Open connection
	db, err = sql.Open("oracle", dsn)
	if err != nil {
		log.Fatal("Error opening DB:", err)
	}

	// Ping DB to verify connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Error connecting to Oracle DB:", err)
	}

	log.Println("Database connected successfully!")
}