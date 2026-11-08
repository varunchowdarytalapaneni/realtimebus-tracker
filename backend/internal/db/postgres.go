package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func ConnectPostgres() *sql.DB {
	connStr := "postgres://bustracker:password123@localhost:5432/bustracker?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to PostgreSQL (PostGIS) successfully!")
	return db
}