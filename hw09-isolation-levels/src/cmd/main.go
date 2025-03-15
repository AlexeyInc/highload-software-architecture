package main

import (
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"hw9-isolation-levels/src/handlers"
	"hw9-isolation-levels/src/storage"
)

func main() {
	registerRoutes()
	storage.InitializeDatabases()

	log.Println("Server running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func registerRoutes() {
	http.HandleFunc("/init-db", handlers.InitDatabaseHandler)
	http.HandleFunc("/lost-update", handlers.HandleLostUpdate)
	http.HandleFunc("/dirty-read", handlers.HandleDirtyRead)
	http.HandleFunc("/non-repeatable-read", handlers.HandleNonRepeatableRead)
	http.HandleFunc("/phantom-read", handlers.HandlePhantomRead)
}
