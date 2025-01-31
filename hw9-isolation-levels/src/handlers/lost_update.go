package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"hw9-isolation-levels/src/storage"
)

func HandleLostUpdate(w http.ResponseWriter, r *http.Request) {
	db, err := storage.GetDbStorage(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req LostUpdateRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	isoLevel := r.URL.Query().Get("isolation")
	if isoLevel == "" {
		http.Error(w, "Missing isolation parameter", http.StatusBadRequest)
		return
	}
	log.Printf("IsolationLevel: %s", isoLevel)

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		transactionA(db.Driver, db.Name, isoLevel, 1, req.ValueA)
	}()
	go func() {
		defer wg.Done()
		transactionB(db.Driver, db.Name, isoLevel, 1, req.ValueB)
	}()

	wg.Wait()

	var finalValue int
	err = db.Driver.QueryRow("SELECT value FROM test_table WHERE id = 1").Scan(&finalValue)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get final value: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Final id = 1 value: %d", finalValue)

	w.Write([]byte(fmt.Sprintf("Lost Update simulation completed. Final value: %d\n", finalValue)))
}

type LostUpdateRequest struct {
	ValueA int `json:"valueA"`
	ValueB int `json:"valueB"`
}

func transactionA(db *sql.DB, driverName, isoLevel string, id, newValue int) {
	err := storage.SetPerconaIsolationLevel(db, driverName, isoLevel)
	if err != nil {
		log.Println("Transaction A failed to set Percona isolation level:", err)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Printf("Transaction A failed to start: %v", err)
		return
	}

	err = storage.SetPostgresIsolationLevel(tx, driverName, isoLevel)
	if err != nil {
		log.Println("Transaction A failed to set Postgres isolation level:", err)
		return
	}

	var currentValue int
	query := storage.FormatQueryPlaceholder("SELECT value FROM test_table WHERE id = ?", driverName)
	if err := tx.QueryRow(query, id).Scan(&currentValue); err != nil {
		log.Printf("Transaction A failed to read value: %v", err)
		tx.Rollback()
		return
	}

	log.Printf("Transaction A for id = 1 got value: %v and adding value: %v", currentValue, newValue)

	time.Sleep(2 * time.Second)

	updateQuery := storage.FormatQueryPlaceholder("UPDATE test_table SET value = ? WHERE id = ?", driverName)
	if _, err := tx.Exec(updateQuery, currentValue+newValue, id); err != nil {
		log.Printf("Transaction A failed to update value: %v", err)
		tx.Rollback()
		return
	}

	tx.Commit()
	log.Println("Transaction A committed.")
}

func transactionB(db *sql.DB, driverName, isoLevel string, id, newValue int) {
	err := storage.SetPerconaIsolationLevel(db, driverName, isoLevel)
	if err != nil {
		log.Println("Transaction B failed to set Percona isolation level:", err)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Printf("Transaction B failed to start: %v", err)
		return
	}

	err = storage.SetPostgresIsolationLevel(tx, driverName, isoLevel)
	if err != nil {
		log.Println("Transaction B failed to set Postgres isolation level:", err)
		return
	}

	time.Sleep(1 * time.Second)

	var currentValue int
	query := storage.FormatQueryPlaceholder("SELECT value FROM test_table WHERE id = ?", driverName)
	if err := tx.QueryRow(query, id).Scan(&currentValue); err != nil {
		log.Printf("Transaction B failed to read value: %v", err)
		tx.Rollback()
		return
	}

	log.Printf("Transaction B for id = 1 got value: %v and adding value: %v", currentValue, newValue)

	updateQuery := storage.FormatQueryPlaceholder("UPDATE test_table SET value = ? WHERE id = ?", driverName)
	if _, err := tx.Exec(updateQuery, currentValue+newValue, id); err != nil {
		log.Printf("Transaction B failed to update value: %v", err)
		tx.Rollback()
		return
	}

	tx.Commit()
	log.Println("Transaction B committed.")
}
