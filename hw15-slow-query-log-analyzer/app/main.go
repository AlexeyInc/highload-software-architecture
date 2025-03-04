package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func initDB() {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s",
		os.Getenv("DATABASE_USER"),
		os.Getenv("DATABASE_PASSWORD"),
		os.Getenv("DATABASE_HOST"),
		os.Getenv("DATABASE_NAME"),
	)
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
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
	rows, err := db.Query(query, "%"+name+"%")
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

	stmt, err := db.Prepare("INSERT INTO users (name, email) VALUES (?, ?)")
	if err != nil {
		log.Fatalf("Failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	for i := 0; i < 1000000; i++ {
		_, err := stmt.Exec(fmt.Sprintf("User%d", i), fmt.Sprintf("user%d@example.com", i))
		if err != nil {
			log.Fatalf("Failed to insert user %d: %v", i, err)
		}
	}
	log.Println("Database populated successfully!")
}

func main() {
	initDB()
	defer db.Close()

	populateDB()

	r := gin.Default()
	r.GET("/slow", slowQuery)
	r.GET("/search", searchUsers)

	port := "5000"
	log.Printf("Server running on port %s", port)
	r.Run(":" + port)
}
