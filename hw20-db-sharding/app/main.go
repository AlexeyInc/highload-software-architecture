package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db *sql.DB

const (
	insertCount = 1000000
	batchSize   = 10
)

type Book struct {
	ID         int    `json:"id"`
	CategoryID int    `json:"category_id"`
	Author     string `json:"author"`
	Title      string `json:"title"`
	Year       int    `json:"year"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using defaults")
	}

	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS books (
		id BIGINT NOT NULL,
		category_id INT NOT NULL,
		author VARCHAR NOT NULL,
		title VARCHAR NOT NULL,
		year INT NOT NULL
	);`)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	http.HandleFunc("/insert", basicInsertHandler)
	http.HandleFunc("/insert-batch", insertBatchHandler)
	http.HandleFunc("/read", readHandler)

	log.Println("Server is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// insertHandler truncates and inserts 1 million new records
func basicInsertHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	_, err := db.Exec("TRUNCATE TABLE books RESTART IDENTITY;")
	if err != nil {
		if !strings.Contains(err.Error(), "is not a table") {
			http.Error(w, "Failed to truncate table", http.StatusInternalServerError)
			return
		}
	}

	err = insertBooks(insertCount)
	if err != nil {
		http.Error(w, "Failed to insert records", http.StatusInternalServerError)
		return
	}

	duration := time.Since(start)
	fmt.Fprintf(w, "Inserted %d records in %v\n", insertCount, duration)
}

// insertBatchHandler inserts 1 thousand new records
func insertBatchHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	err := insertBooks(batchSize)
	if err != nil {
		http.Error(w, "Failed to insert records", http.StatusInternalServerError)
		return
	}

	duration := time.Since(start)
	fmt.Fprintf(w, "Inserted %d records in %v\n", batchSize, duration)
}

// insertBooks inserts N records in a batch
func insertBooks(count int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	valueStrings := make([]string, 0, 100)
	valueArgs := make([]any, 0, 100*4)
	stmtTemplate := "INSERT INTO books (id, category_id, author, title, year) VALUES %s"

	for i := range count {
		categoryID := randomCategory() // Ensure it falls within shard criteria
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", len(valueArgs)+1, len(valueArgs)+2, len(valueArgs)+3, len(valueArgs)+4, len(valueArgs)+5))
		valueArgs = append(valueArgs, i, categoryID, randomAuthor(), fmt.Sprintf("Book %d", i), randomYear())

		if (i+1)%100 == 0 || i+1 == count {
			stmt := fmt.Sprintf(stmtTemplate, join(valueStrings, ","))
			_, err := tx.Exec(stmt, valueArgs...)
			if err != nil {
				return err
			}
			valueStrings = valueStrings[:0]
			valueArgs = valueArgs[:0]
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Helper function to join slice elements into a single string
func join(elements []string, sep string) string {
	result := ""
	for i, element := range elements {
		if i > 0 {
			result += sep
		}
		result += element
	}
	return result
}

// readHandler fetches 1000 random records
func readHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var totalCount int
	err := db.QueryRow("SELECT COUNT(*) FROM books").Scan(&totalCount)
	if err != nil || totalCount == 0 {
		http.Error(w, "Failed to count records", http.StatusInternalServerError)
		return
	}

	offset := rand.Intn(totalCount - batchSize + 1)

	rows, err := db.Query("SELECT id, category_id, author, title, year FROM books LIMIT $1 OFFSET $2", batchSize, offset)
	if err != nil {
		http.Error(w, "Failed to read records", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		if err := rows.Scan(&book.ID, &book.CategoryID, &book.Author, &book.Title, &book.Year); err != nil {
			http.Error(w, "Failed to scan record", http.StatusInternalServerError)
			return
		}
		books = append(books, book)
	}

	duration := time.Since(start)
	response := map[string]any{
		"count":    len(books),
		"duration": duration.String(),
		"books":    books,
	}

	json.NewEncoder(w).Encode(response)
}

// randomAuthor generates a random author name
func randomAuthor() string {
	authors := []string{"J.K. Rowling", "George Orwell", "Timoti Liri", "Mark Twain", "F. Scott Fitzgerald"}
	return authors[rand.Intn(len(authors))]
}

// randomYear generates a random year between 1900 and 2023
func randomYear() int {
	return rand.Intn(124) + 1900
}

func randomCategory() int {
	categories := []int{1, 2, 3, 4}
	return categories[rand.Intn(len(categories))]
}
