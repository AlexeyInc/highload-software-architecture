package storage

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
)

var (
	PostgresDB *DB
	PerconaDB  *DB
)

type DB struct {
	Driver *sql.DB
	Name   string
}

// Database initialization
func InitializeDatabases() {
	PerconaDB = connectToDB("mysql", "testuser:testpassword@tcp(127.0.0.1:3306)/testdb")
	PostgresDB = connectToDB("postgres", "postgres://testuser:testpassword@localhost:5432/testdb?sslmode=disable")
}

func connectToDB(driver, dsn string) *DB {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		log.Fatalf("Failed to connect to %s: %v", driver, err)
	}

	name := "percona"
	if driver == "postgres" {
		name = "postgres"
	}
	return &DB{
		Driver: db,
		Name:   name,
	}
}

func GetDbStorage(r *http.Request) (*DB, error) {
	dbType := r.URL.Query().Get("dbDriver")
	switch dbType {
	case "percona":
		return PerconaDB, nil
	case "postgres":
		return PostgresDB, nil
	default:
		return nil, fmt.Errorf("Invalid dbDriver parameter. Use 'percona' or 'postgres'")
	}
}
