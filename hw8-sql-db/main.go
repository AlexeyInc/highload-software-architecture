package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func main() {
	initializeDatabase()

	registerRoutes()

	log.Println("Server running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initializeDatabase() {
	var err error
	db, err = sql.Open("mysql", "testuser:testpassword@tcp(127.0.0.1:3306)/testdb")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	query := `CREATE TABLE IF NOT EXISTS users (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100),
		email VARCHAR(100),
		date_of_birth DATE
	) ENGINE=InnoDB;`
	if _, err = db.Exec(query); err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	fmt.Println("Database initialized and table created.")
}

func registerRoutes() {
	http.HandleFunc("/insertUsers", insertUsersHandler)
	http.HandleFunc("/measureSelectPerformance/withoutIndex", func(w http.ResponseWriter, r *http.Request) {
		measureQueryPerformance(w, "SELECT * FROM users WHERE date_of_birth BETWEEN '1990-01-01' AND '1991-01-01' LIMIT 1000")
	})
	http.HandleFunc("/measureSelectPerformance/withBTREE", func(w http.ResponseWriter, r *http.Request) {
		measureQueryPerformance(w, "SELECT * FROM users WHERE date_of_birth BETWEEN '1990-01-01' AND '1991-01-01' LIMIT 1000")
	})
	http.HandleFunc("/measureSelectPerformance/withHASH", func(w http.ResponseWriter, r *http.Request) {
		measureQueryPerformance(w, "SELECT * FROM users WHERE date_of_birth BETWEEN '1990-01-01' AND '1991-01-01' LIMIT 1000")
	})
	http.HandleFunc("/changeFlushLogSetting", changeFlushLogSettingHandler)
	http.HandleFunc("/insertUser", insertSingleUserHandler)
	http.HandleFunc("/manageIndex", manageIndexHandler) // New endpoint for creating/deleting an index
}

func insertUsersHandler(w http.ResponseWriter, r *http.Request) {
	const (
		batchSize = 10000
		total     = 40_000_000
	)
	start := time.Now()

	for i := 0; i < total/batchSize; i++ {
		queryBuilder := strings.Builder{}
		queryBuilder.WriteString("INSERT INTO users (name, email, date_of_birth) VALUES ")
		values := make([]interface{}, 0, batchSize*3)
		for j := 0; j < batchSize; j++ {
			name := fmt.Sprintf("User%d", rand.Intn(total))
			email := fmt.Sprintf("user%d@example.com", rand.Intn(total))
			dob := time.Date(rand.Intn(40)+1980, time.Month(rand.Intn(12)+1), rand.Intn(28)+1, 0, 0, 0, 0, time.UTC)
			queryBuilder.WriteString("(?, ?, ?),")
			values = append(values, name, email, dob)
		}
		query := queryBuilder.String()
		query = query[:len(query)-1] // Remove trailing comma
		if _, err := db.Exec(query, values...); err != nil {
			http.Error(w, fmt.Sprintf("Failed to insert batch: %v", err), http.StatusInternalServerError)
			return
		}
		fmt.Printf("\rInserted %d/%d users", (i+1)*batchSize, total)
	}

	fmt.Printf("\nInsertion completed in %v\n", time.Since(start))
	w.Write([]byte("40M users inserted successfully"))
}

func measureQueryPerformance(w http.ResponseWriter, query string) {
	start := time.Now()
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to execute query: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	duration := time.Since(start)

	resMsg := fmt.Sprintf("\nQuery completed in %v, rows fetched: %d", duration, count)
	fmt.Print(resMsg)
	w.Write([]byte(resMsg))
}

func changeFlushLogSettingHandler(w http.ResponseWriter, r *http.Request) {
	value := r.URL.Query().Get("innodb_flush_log_at_trx_commit")
	flushLog, err := strconv.Atoi(value)
	if err != nil || (flushLog != 0 && flushLog != 1 && flushLog != 2) {
		http.Error(w, "Invalid value for innodb_flush_log_at_trx_commit. Allowed values: 0, 1, 2.", http.StatusBadRequest)
		return
	}

	dbAdmin, err := sql.Open("mysql", "root:rootpassword@tcp(127.0.0.1:3306)/")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect as root: %v", err), http.StatusInternalServerError)
		return
	}
	defer dbAdmin.Close()

	if _, err := dbAdmin.Exec(fmt.Sprintf("SET GLOBAL innodb_flush_log_at_trx_commit = %d;", flushLog)); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update innodb_flush_log_at_trx_commit: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(fmt.Sprintf("innodb_flush_log_at_trx_commit set to %d", flushLog)))
}

func insertSingleUserHandler(w http.ResponseWriter, r *http.Request) {
	name := fmt.Sprintf("User%d", rand.Intn(1000000))
	email := fmt.Sprintf("user%d@example.com", rand.Intn(1000000))
	dob := time.Date(rand.Intn(50)+1970, time.Month(rand.Intn(12)+1), rand.Intn(28)+1, 0, 0, 0, 0, time.UTC)

	if _, err := db.Exec("INSERT INTO users (name, email, date_of_birth) VALUES (?, ?, ?);", name, email, dob); err != nil {
		http.Error(w, fmt.Sprintf("Failed to insert user: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("User inserted successfully"))
}

func manageIndexHandler(w http.ResponseWriter, r *http.Request) {
	indexType := r.URL.Query().Get("indexType")
	action := r.URL.Query().Get("action")

	// Validate parameters
	if indexType == "" || (indexType != "BTREE" && indexType != "HASH") {
		http.Error(w, "Invalid or missing indexType. Allowed values: BTREE, HASH.", http.StatusBadRequest)
		return
	}
	if action == "" || (action != "create" && action != "delete") {
		http.Error(w, "Invalid or missing action. Allowed values: create, delete.", http.StatusBadRequest)
		return
	}

	indexName := fmt.Sprintf("idx_dob_%s", strings.ToLower(indexType))

	switch action {
	case "create":
		if _, err := db.Exec(fmt.Sprintf("CREATE INDEX %s ON users(date_of_birth) USING %s;", indexName, indexType)); err != nil {
			http.Error(w, fmt.Sprintf("Failed to create %s index: %v", indexType, err), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(fmt.Sprintf("%s index created successfully.", indexType)))
	case "delete":
		if _, err := db.Exec(fmt.Sprintf("DROP INDEX %s ON users;", indexName)); err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete %s index: %v", indexType, err), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(fmt.Sprintf("%s index deleted successfully.", indexType)))
	}
}
