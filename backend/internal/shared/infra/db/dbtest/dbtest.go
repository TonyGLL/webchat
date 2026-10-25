// Package dbtest gives integration tests a throwaway, fully migrated Postgres database.
package dbtest

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"backend/migrations"

	_ "github.com/lib/pq"
)

// New creates a fresh database next to the one in TEST_DATABASE_URL, applies every migration to
// it and drops it when the test ends. The test is skipped when TEST_DATABASE_URL is not set.
func New(t *testing.T) *sql.DB {
	t.Helper()
	base := os.Getenv("TEST_DATABASE_URL")
	if base == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	admin, err := sql.Open("postgres", base)
	if err != nil {
		t.Fatal(err)
	}
	name := fmt.Sprintf("dbtest_%d", time.Now().UnixNano())
	if _, err := admin.Exec(`CREATE DATABASE ` + name); err != nil {
		admin.Close()
		t.Fatalf("create database: %v", err)
	}

	u, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	u.Path = "/" + strings.TrimPrefix(name, "/")
	db, err := sql.Open("postgres", u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
		_, _ = admin.Exec(`DROP DATABASE IF EXISTS ` + name + ` WITH (FORCE)`)
		admin.Close()
	})
	if err := migrations.Up(db); err != nil {
		t.Fatalf("migrations: %v", err)
	}
	return db
}
