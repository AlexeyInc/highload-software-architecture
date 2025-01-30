package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"hw9-isolation-levels/src/storage"
)

func HandleNonRepeatableRead(w http.ResponseWriter, r *http.Request) {
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
		transactionANonRepeatableReader(db.Driver, db.Name, 1)
	}()
	go func() {
		defer wg.Done()
		transactionBNonRepeatableWriter(db.Driver, db.Name, 1, newValue)
	}()

	wg.Wait()
	w.Write([]byte("Non-Repeatable Read simulation completed. Check logs."))
}

func transactionANonRepeatableReader(db *sql.DB, driverName string, id int) {
	tx, err := db.Begin()
	if err != nil {
		log.Println("Transaction A failed to start:", err)
		return
	}
	defer tx.Commit()

	var value1, value2 int
	query := storage.FormatQueryPlaceholder("SELECT value FROM test_table WHERE id = ?", driverName)
	err = tx.QueryRow(query, id).Scan(&value1)
	if err != nil {
		log.Println("Transaction A failed to read first value:", err)
		return
	}
	log.Printf("Transaction A first read: %d", value1)

	time.Sleep(2 * time.Second) // Allow writer to update value

	err = tx.QueryRow(query, id).Scan(&value2)
	if err != nil {
		log.Println("Transaction A failed to read second value:", err)
		return
	}
	log.Printf("Transaction A second read: %d", value2)
}

func transactionBNonRepeatableWriter(db *sql.DB, driverName string, id, newValue int) {
	time.Sleep(300 * time.Millisecond) // Ensure A reads first

	tx, err := db.Begin()
	if err != nil {
		log.Println("Transaction B failed to start:", err)
		return
	}

	query := storage.FormatQueryPlaceholder("UPDATE test_table SET value = ? WHERE id = ?", driverName)
	_, err = tx.Exec(query, newValue, id)
	if err != nil {
		log.Println("Transaction B failed to update:", err)
		tx.Rollback()
		return
	}

	tx.Commit()
	log.Println("Transaction B committed update.")
}
