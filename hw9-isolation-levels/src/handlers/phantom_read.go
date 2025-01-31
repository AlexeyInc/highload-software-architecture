package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"sync"
	"time"

	"hw9-isolation-levels/src/storage"
)

func HandlePhantomRead(w http.ResponseWriter, r *http.Request) {
	db, err := storage.GetDbStorage(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	isoLevel := r.URL.Query().Get("isolation")
	if isoLevel == "" {
		http.Error(w, "Missing isolation parameter", http.StatusBadRequest)
		return
	}
	log.Printf("IsolationLevel: %s", isoLevel)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		transactionAPhantomReader(db.Driver, db.Name, isoLevel)
	}()
	go func() {
		defer wg.Done()
		transactionBPhantomWriter(db.Driver, db.Name, isoLevel)
	}()

	wg.Wait()

	w.Write([]byte("Phantom Read simulation completed. Check logs.\n"))
}

func transactionAPhantomReader(db *sql.DB, driverName, isoLevel string) {
	err := storage.SetPerconaIsolationLevel(db, driverName, isoLevel)
	if err != nil {
		log.Println("Transaction for PhantomRead failed to set Percona isolation level:", err)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Println("Transaction A failed to start:", err)
		return
	}
	defer tx.Commit()

	err = storage.SetPostgresIsolationLevel(tx, driverName, isoLevel)
	if err != nil {
		log.Println("Transaction A PhantomReader failed to set Postgres isolation level:", err)
		return
	}

	var count1, count2 int
	query := storage.FormatQueryPlaceholder("SELECT COUNT(*) FROM test_table WHERE value > 10", driverName)
	err = tx.QueryRow(query).Scan(&count1)
	if err != nil {
		log.Println("Transaction A failed to read first count:", err)
		return
	}
	log.Printf("Transaction A first count: %d", count1)

	time.Sleep(2 * time.Second) // Allow writer to insert new rows

	err = tx.QueryRow(query).Scan(&count2)
	if err != nil {
		log.Println("Transaction A failed to read second count:", err)
		return
	}
	log.Printf("Transaction A second count: %d", count2)
}

func transactionBPhantomWriter(db *sql.DB, driverName, isoLevel string) {
	time.Sleep(time.Second) // Ensure A reads first

	tx, err := db.Begin()
	if err != nil {
		log.Println("Transaction B failed to start:", err)
		return
	}
	defer tx.Commit()

	err = storage.SetPostgresIsolationLevel(tx, driverName, isoLevel)
	if err != nil {
		log.Println("Transaction B failed to set Postgres isolation level:", err)
		return
	}

	query := storage.FormatQueryPlaceholder("INSERT INTO test_table (value) VALUES (?)", driverName)
	_, err = tx.Exec(query, 999)
	if err != nil {
		log.Println("Transaction B failed to insert:", err)
		tx.Rollback()
		return
	}

	log.Println("Transaction B inserted new row.")
}
