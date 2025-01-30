package handlers

import (
	"fmt"
	"net/http"

	"hw9-isolation-levels/src/storage"
)

func InitDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	db, err := storage.GetDbStorage(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var createTableQuery, insertDataQuery string

	if db.Name == "percona" {
		createTableQuery = `
		CREATE TABLE IF NOT EXISTS test_table (
			id INT AUTO_INCREMENT PRIMARY KEY,
			value INT NOT NULL
		);`
		insertDataQuery = "INSERT IGNORE INTO test_table (id, value) VALUES (1, 10);"
	} else {
		createTableQuery = `
		CREATE TABLE IF NOT EXISTS test_table (
			id SERIAL PRIMARY KEY,
			value INT NOT NULL
		);`
		insertDataQuery = "INSERT INTO test_table (id, value) VALUES (1, 10) ON CONFLICT (id) DO NOTHING;"
	}

	_, err = db.Driver.Exec(createTableQuery)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create table: %v", err), http.StatusInternalServerError)
		return
	}

	_, err = db.Driver.Exec(insertDataQuery)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to insert initial data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Database initialized successfully."))
}
