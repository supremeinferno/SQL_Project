package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/sijms/go-ora/v2"
)

const defaultOracleDSN = "oracle://system:Oracle123@localhost:1521/FREEPDB1"

var db *sql.DB

func connectDB() {
	dsn := os.Getenv("ORACLE_DSN")
	if dsn == "" {
		dsn = defaultOracleDSN
	}

	var err error
	db, err = sql.Open("oracle", dsn)
	if err != nil {
		log.Fatalf("open Oracle connection: %v", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err = db.Ping(); err != nil {
		log.Fatalf("connect to Oracle database: %v", err)
	}

	log.Println("Database connected successfully")
}
