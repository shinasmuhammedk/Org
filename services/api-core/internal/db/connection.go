package db

import (
	"database/sql"
	"log"
    
    _ "github.com/lib/pq"
)

var DB *sql.DB
var QueriesInstance *Queries

func Init() {
	connStr := "postgres://postgres:Shinas@localhost:5432/Org_db?sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("DB not connected")
	}

	DB = db
	QueriesInstance = New(db)
}
