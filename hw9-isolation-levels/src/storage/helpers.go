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

func SetIsolationLevel(db *sql.DB, driverName, isoLevel string) error {
	validLevels := map[string]bool{
		"READ UNCOMMITTED": true,
		"READ COMMITTED":   true,
		"REPEATABLE READ":  true,
		"SERIALIZABLE":     true,
	}

	if _, ok := validLevels[isoLevel]; !ok {
		return fmt.Errorf("invalid isolation level: %s", isoLevel)
	}

	var query string
	switch driverName {
	case "percona", "mysql":
		query = fmt.Sprintf("SET SESSION TRANSACTION ISOLATION LEVEL %s;", isoLevel)
	case "postgres":
		query = fmt.Sprintf("SET TRANSACTION ISOLATION LEVEL %s;", isoLevel)
	default:
		return fmt.Errorf("unsupported database driver: %s", driverName)
	}

	if driverName == "percona" {
		_, err := db.Exec(query)
		if err != nil {
			log.Printf("Failed to set isolation level to %s for %s: %v", isoLevel, driverName, err)
			return err
		}
	}

	return nil
}

func SetPerconaIsolationLevel(db *sql.DB, driverName, isoLevel string) error {
	if driverName == "percona" { // TODO: replace to constant
		query := fmt.Sprintf("SET TRANSACTION ISOLATION LEVEL %s;", isoLevel)
		_, err := db.Exec(query)
		if err != nil {
			log.Printf("Failed to set isolation level to %s for %s: %v", isoLevel, driverName, err)
			return err
		}
	}

	return nil
}

func SetPostgresIsolationLevel(tx *sql.Tx, driverName, isoLevel string) error {
	if driverName == "postgres" {
		_, err := tx.Exec(fmt.Sprintf("SET TRANSACTION ISOLATION LEVEL %s", isoLevel))
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return nil
}
