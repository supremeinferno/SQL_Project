package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/sijms/go-ora/v2"
)

var db *sql.DB

func connectDB() {
	dsn := os.Getenv("ORACLE_DSN")
	if dsn == "" {
		dsn = "oracle://admin:password@localhost:1521/ORCLPDB1"
	}

	var err error
	db, err = sql.Open("oracle", dsn)
	if err != nil {
		panic(err)
	}

	// Ping the database to verify the connection
	if err = db.Ping(); err != nil {
		log.Printf("Warning: Could not ping Oracle DB: %v", err)
	}
}
