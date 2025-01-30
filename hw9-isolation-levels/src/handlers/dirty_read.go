package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"hw9-isolation-levels/src/storage"
)

func HandleDirtyRead(w http.ResponseWriter, r *http.Request) {
	db, err := storage.GetDbStorage(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newValueStr := r.URL.Query().Get("newValue")
	if newValueStr == "" {
		http.Error(w, "Missing newValue parameter", http.StatusBadRequest)
		return
	}

	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		http.Error(w, "Invalid newValue parameter", http.StatusBadRequest)
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		transactionADirtyReadWriter(db.Driver, db.Name, 1, newValue)
	}()
	go func() {
		defer wg.Done()
		transactionBDirtyReadReader(db.Driver, db.Name, 1)
	}()

	wg.Wait()

	var finalValue int
	err = db.Driver.QueryRow("SELECT value FROM test_table WHERE id = 1").Scan(&finalValue)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get final value: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Final id = 1 value: %d", finalValue)

	w.Write([]byte("Dirty Read simulation completed. Check logs."))
}
func transactionADirtyReadWriter(db *sql.DB, driverName string, id, newValue int) {
	tx, err := db.Begin()
	if err != nil {
		log.Println("Transaction A failed to start:", err)
		return
	}

	query := storage.FormatQueryPlaceholder("UPDATE test_table SET value = ? WHERE id = ?", driverName)
	_, err = tx.Exec(query, newValue, id)
	if err != nil {
		log.Println("Transaction A failed to update:", err)
		tx.Rollback()
		return
	}

	log.Printf("Transaction A updated id = %v value to %v, but not committed yet.", id, newValue)
	time.Sleep(5 * time.Second)

	tx.Rollback()
	log.Println("Transaction A rolled back.")
}
func transactionBDirtyReadReader(db *sql.DB, driverName string, id int) {
	time.Sleep(2 * time.Second) // Ensure A updates before B reads

	tx, err := db.Begin()
	if err != nil {
		log.Println("Transaction B failed to start:", err)
		return
	}
	defer tx.Commit()

	var value int
	query := storage.FormatQueryPlaceholder("SELECT value FROM test_table WHERE id = ?", driverName)
	err = tx.QueryRow(query, id).Scan(&value)
	if err != nil {
		log.Println("Transaction B failed to read:", err)
		return
	}

	log.Printf("Transaction B read value: %d", value)
}
