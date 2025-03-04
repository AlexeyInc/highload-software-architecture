package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func initDB() {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("DATABASE_USER"),
		os.Getenv("DATABASE_PASSWORD"),
		os.Getenv("DATABASE_HOST"),
		os.Getenv("DATABASE_PORT"),
		os.Getenv("DATABASE_NAME"),
	)
	maxRetries := 3
	for retries := range maxRetries {
		db, err = sql.Open("mysql", dsn)
		if err == nil {
			if err = db.Ping(); err == nil {
				log.Println("Connected to MySQL successfully")
				break
			}
		}
		log.Printf("Waiting for MySQL to be ready... retry %d/%d. user: %s, S: %s", retries+1, maxRetries, os.Getenv("DATABASE_USER"), os.Getenv("DATABASE_HOST"))
		time.Sleep(3 * time.Second) // Wait before retrying
	}
	if err != nil {
		log.Fatalf("Failed to connect to MySQL after retries: %v", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0)
}

func slowQuery(c *gin.Context) {
	timeout := c.DefaultQuery("timeout", "2")
	query := "SELECT SLEEP(?)"
	_, err := db.Exec(query, timeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Query executed successfully"})
}

func searchUsers(c *gin.Context) {
	name := c.DefaultQuery("name", "")
	query := "SELECT id, name, email FROM users WHERE name LIKE ?"
	rows, err := db.Query(query, name+"%")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []map[string]any
	for rows.Next() {
		var id int
		var name, email string
		if err := rows.Scan(&id, &name, &email); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		users = append(users, map[string]any{"id": id, "name": name, "email": email})
	}

	c.JSON(http.StatusOK, users)
}

func populateDB() {
	_, err := db.Exec("DROP TABLE IF EXISTS users")
	if err != nil {
		log.Fatalf("Failed to drop table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100),
			email VARCHAR(100)
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	const batchSize = 10000
	values := make([]string, 0, batchSize)
	args := make([]any, 0, batchSize*2) // Two placeholders per row (name, email)

	log.Println("Starting database population...")

	for i := 1; i <= 1000000; i++ {
		values = append(values, "(?, ?)")
		args = append(args, fmt.Sprintf("User%d", i), fmt.Sprintf("user%d@example.com", i))

		// When we reach batch size, execute the batch insert
		if i%batchSize == 0 {
			insertQuery := fmt.Sprintf("INSERT INTO users (name, email) VALUES %s", strings.Join(values, ","))
			_, err := db.Exec(insertQuery, args...)
			if err != nil {
				log.Fatalf("Failed to insert batch: %v", err)
			}

			values = values[:0]
			args = args[:0]

			log.Printf("Inserted %d records...", i)
		}
	}

	// Insert any remaining records that didn't fit in the last batch
	if len(values) > 0 {
		insertQuery := fmt.Sprintf("INSERT INTO users (name, email) VALUES %s", strings.Join(values, ","))
		_, err := db.Exec(insertQuery, args...)
		if err != nil {
			log.Fatalf("Failed to insert final batch: %v", err)
		}
	}

	log.Println("Database populated successfully!")
}

func main() {
	initDB()
	defer db.Close()

	populateDB()

	r := gin.Default()
	r.GET("/query", slowQuery)
	r.GET("/search", searchUsers)

	port := "8080"
	log.Printf("Server running on port %s", port)
	r.Run(":" + port)
}
