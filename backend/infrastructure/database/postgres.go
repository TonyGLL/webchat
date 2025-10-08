package database

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib" // Standard library database driver for pgx
)

// NewDBPool creates a new database connection pool using the provided connection string.
func NewDBPool(connString string) (*sql.DB, error) {
	if connString == "" {
		return nil, fmt.Errorf("database connection string is empty")
	}

	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("unable to open database connection: %w", err)
	}

	// Ping the database to verify the connection is alive.
	err = db.Ping()
	if err != nil {
		db.Close() // Close the connection if ping fails
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return db, nil
}
