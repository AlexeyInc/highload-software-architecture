package storage

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
)

func FormatQueryPlaceholder(query string, driverName string) string {
	if driverName == "postgres" {
		var counter = 0
		return regexp.MustCompile(`\?`).ReplaceAllStringFunc(query, func(_ string) string {
			counter++
			return fmt.Sprintf("$%d", counter)
		})
	}
	return query
}

func SetIsolationLevel(tx *sql.Tx, isoLevel string) {
	query := fmt.Sprintf("SET TRANSACTION ISOLATION LEVEL %s;", isoLevel)
	_, err := tx.Exec(query)
	if err != nil {
		log.Printf("Failed to set isolation level to %s: %v", isoLevel, err)
	}
}
