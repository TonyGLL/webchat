package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // registers the "postgres" driver used below
)

// PoolSettings tunes the database/sql connection pool.
type PoolSettings struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// NewDBPool creates a new database connection pool using the provided connection string.
func NewDBPool(connString string, pool PoolSettings) (*sql.DB, error) {
	if connString == "" {
		return nil, fmt.Errorf("database connection string is empty")
	}

	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("unable to open database connection: %w", err)
	}

	db.SetMaxOpenConns(pool.MaxOpenConns)
	db.SetMaxIdleConns(pool.MaxIdleConns)
	db.SetConnMaxLifetime(pool.ConnMaxLifetime)
	db.SetConnMaxIdleTime(pool.ConnMaxIdleTime)

	// Ping the database to verify the connection is alive.
	err = db.Ping()
	if err != nil {
		db.Close() // Close the connection if ping fails
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return db, nil
}
