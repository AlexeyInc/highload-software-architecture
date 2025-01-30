package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

var (
	postgresDB *DB
	perconaDB  *DB
)

type DB struct {
	db   *sql.DB
	name string
}

func main() {
	initializeDatabases()
	registerRoutes()

	log.Println("Server running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Database initialization
func initializeDatabases() {
	perconaDB = connectToDB("mysql", "testuser:testpassword@tcp(127.0.0.1:3306)/testdb")
	postgresDB = connectToDB("postgres", "postgres://testuser:testpassword@localhost:5432/testdb?sslmode=disable")
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
		db:   db,
		name: name,
	}
}

func registerRoutes() {
	http.HandleFunc("/init-db", initDatabaseHandler)
	http.HandleFunc("/lost-update", handleLostUpdate)
	http.HandleFunc("/dirty-read", handleDirtyRead)
	http.HandleFunc("/non-repeatable-read", handleNonRepeatableRead)
	http.HandleFunc("/phantom-read", handlePhantomRead)
}

func initDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	storage, err := getDbStorage(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var createTableQuery, insertDataQuery string

	if storage.name == "percona" {
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

	_, err = storage.db.Exec(createTableQuery)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create table: %v", err), http.StatusInternalServerError)
		return
	}

	_, err = storage.db.Exec(insertDataQuery)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to insert initial data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Database initialized successfully."))
}

// === LOST UPDATE SIMULATION ===
func handleLostUpdate(w http.ResponseWriter, r *http.Request) {
	storage, err := getDbStorage(r)
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

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		transactionA(storage.db, storage.name, 1, req.ValueA)
	}()
	go func() {
		defer wg.Done()
		transactionB(storage.db, storage.name, 1, req.ValueB)
	}()

	wg.Wait()

	var finalValue int
	err = storage.db.QueryRow("SELECT value FROM test_table WHERE id = 1").Scan(&finalValue)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get final value: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Final id = 1 value: %d", finalValue)
	w.Write([]byte(fmt.Sprintf("Lost Update simulation completed. Final value: %d", finalValue)))
}

type LostUpdateRequest struct {
	ValueA int `json:"valueA"`
	ValueB int `json:"valueB"`
}

func transactionA(db *sql.DB, driverName string, id, newValue int) {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Transaction A failed to start: %v", err)
		return
	}

	var currentValue int
	query := formatQueryPlaceholder("SELECT value FROM test_table WHERE id = ?", driverName)
	if err := tx.QueryRow(query, id).Scan(&currentValue); err != nil {
		log.Printf("Transaction A failed to read value: %v", err)
		tx.Rollback()
		return
	}

	log.Printf("Transaction A for id = 1 got value: %v and adding value: %v", currentValue, newValue)

	time.Sleep(2 * time.Second)

	updateQuery := formatQueryPlaceholder("UPDATE test_table SET value = ? WHERE id = ?", driverName)
	if _, err := tx.Exec(updateQuery, currentValue+newValue, id); err != nil {
		log.Printf("Transaction A failed to update value: %v", err)
		tx.Rollback()
		return
	}

	tx.Commit()
	log.Println("Transaction A committed.")
}

func transactionB(db *sql.DB, driverName string, id, newValue int) {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Transaction B failed to start: %v", err)
		return
	}

	time.Sleep(1 * time.Second)

	var currentValue int
	query := formatQueryPlaceholder("SELECT value FROM test_table WHERE id = ?", driverName)
	if err := tx.QueryRow(query, id).Scan(&currentValue); err != nil {
		log.Printf("Transaction B failed to read value: %v", err)
		tx.Rollback()
		return
	}

	log.Printf("Transaction B for id = 1 got value: %v and adding value: %v", currentValue, newValue)

	updateQuery := formatQueryPlaceholder("UPDATE test_table SET value = ? WHERE id = ?", driverName)
	if _, err := tx.Exec(updateQuery, currentValue+newValue, id); err != nil {
		log.Printf("Transaction B failed to update value: %v", err)
		tx.Rollback()
		return
	}

	tx.Commit()
	log.Println("Transaction B committed.")
}

// === DIRTY READ SIMULATION ===
func handleDirtyRead(w http.ResponseWriter, r *http.Request) {
	storage, err := getDbStorage(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		transactionDirtyReadWriter(storage.db, storage.name, 1, 200)
	}()
	go func() {
		defer wg.Done()
		transactionDirtyReadReader(storage.db, storage.name, 1)
	}()

	wg.Wait()
	w.Write([]byte("Dirty Read simulation completed. Check logs."))
}

func transactionDirtyReadWriter(db *sql.DB, driverName string, id, newValue int) {
	tx, err := db.Begin()
	if err != nil {
		log.Println("Transaction A failed to start:", err)
		return
	}

	query := formatQueryPlaceholder("UPDATE test_table SET value = ? WHERE id = ?", driverName)
	_, err = tx.Exec(query, newValue, id)
	if err != nil {
		log.Println("Transaction A failed to update:", err)
		tx.Rollback()
		return
	}

	log.Println("Transaction A updated but not committed yet.")
	time.Sleep(1 * time.Second)

	tx.Rollback()
	log.Println("Transaction A rolled back.")
}

func transactionDirtyReadReader(db *sql.DB, driverName string, id int) {
	time.Sleep(300 * time.Millisecond) // Ensure A updates before B reads

	tx, err := db.Begin()
	if err != nil {
		log.Println("Transaction B failed to start:", err)
		return
	}
	defer tx.Commit()

	var value int
	query := formatQueryPlaceholder("SELECT value FROM test_table WHERE id = ?", driverName)
	err = tx.QueryRow(query, id).Scan(&value)
	if err != nil {
		log.Println("Transaction B failed to read:", err)
		return
	}

	log.Printf("Transaction B read value: %d", value)
}

// === NON-REPEATABLE READ SIMULATION ===
func handleNonRepeatableRead(w http.ResponseWriter, r *http.Request) {
	storage, err := getDbStorage(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		transactionNonRepeatableReader(storage.db, storage.name, 1)
	}()
	go func() {
		defer wg.Done()
		transactionNonRepeatableWriter(storage.db, storage.name, 1, 300)
	}()

	wg.Wait()
	w.Write([]byte("Non-Repeatable Read simulation completed. Check logs."))
}

func transactionNonRepeatableReader(db *sql.DB, driverName string, id int) {
	tx, err := db.Begin()
	if err != nil {
		log.Println("Transaction A failed to start:", err)
		return
	}
	defer tx.Commit()

	var value1, value2 int
	query := formatQueryPlaceholder("SELECT value FROM test_table WHERE id = ?", driverName)
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

func transactionNonRepeatableWriter(db *sql.DB, driverName string, id, newValue int) {
	time.Sleep(500 * time.Millisecond) // Ensure A reads first

	tx, err := db.Begin()
	if err != nil {
		log.Println("Transaction B failed to start:", err)
		return
	}

	query := formatQueryPlaceholder("UPDATE test_table SET value = ? WHERE id = ?", driverName)
	_, err = tx.Exec(query, newValue, id)
	if err != nil {
		log.Println("Transaction B failed to update:", err)
		tx.Rollback()
		return
	}

	tx.Commit()
	log.Println("Transaction B committed update.")
}

// === PHANTOM READ SIMULATION ===
func handlePhantomRead(w http.ResponseWriter, r *http.Request) {
	storage, err := getDbStorage(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		transactionPhantomReader(storage.db, storage.name)
	}()
	go func() {
		defer wg.Done()
		transactionPhantomWriter(storage.db, storage.name)
	}()

	wg.Wait()
	w.Write([]byte("Phantom Read simulation completed. Check logs."))
}

func transactionPhantomReader(db *sql.DB, driverName string) {
	tx, err := db.Begin()
	if err != nil {
		log.Println("Transaction A failed to start:", err)
		return
	}
	defer tx.Commit()

	var count1, count2 int
	query := formatQueryPlaceholder("SELECT COUNT(*) FROM test_table WHERE value > 10", driverName)
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

func transactionPhantomWriter(db *sql.DB, driverName string) {
	time.Sleep(500 * time.Millisecond) // Ensure A reads first

	tx, err := db.Begin()
	if err != nil {
		log.Println("Transaction B failed to start:", err)
		return
	}

	query := formatQueryPlaceholder("INSERT INTO test_table (value) VALUES (?)", driverName)
	_, err = tx.Exec(query, 999)
	if err != nil {
		log.Println("Transaction B failed to insert:", err)
		tx.Rollback()
		return
	}

	tx.Commit()
	log.Println("Transaction B inserted new row.")
}

// Helper functions
func getDbStorage(r *http.Request) (*DB, error) {
	dbType := r.URL.Query().Get("dbDriver")
	switch dbType {
	case "percona":
		return perconaDB, nil
	case "postgres":
		return postgresDB, nil
	default:
		return nil, fmt.Errorf("Invalid dbDriver parameter. Use 'percona' or 'postgres'")
	}
}

func formatQueryPlaceholder(query string, driverName string) string {
	if driverName == "postgres" {
		var counter = 0
		return regexp.MustCompile(`\?`).ReplaceAllStringFunc(query, func(_ string) string {
			counter++
			return fmt.Sprintf("$%d", counter)
		})
	}
	return query
}
