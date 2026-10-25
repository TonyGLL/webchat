// Package dbtest gives repository tests a migrated, throwaway PostgreSQL database.
package dbtest

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"testing"

	"backend/migrations"

	_ "github.com/lib/pq"
)

// New creates a fresh database next to the one TEST_DATABASE_URL points at, applies every
// migration to it and drops it when the test ends. The test is skipped when the variable is
// not set, so `go test ./...` works without a database.
func New(t *testing.T) *sql.DB {
	t.Helper()
	raw := os.Getenv("TEST_DATABASE_URL")
	if raw == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	admin, err := sql.Open("postgres", raw)
	if err != nil {
		t.Fatal(err)
	}
	name := fmt.Sprintf("webchat_test_%d", rand.Int63())
	if _, err := admin.Exec("CREATE DATABASE " + name); err != nil {
		admin.Close()
		t.Fatalf("cannot create the test database: %v", err)
	}

	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	u.Path = "/" + name
	db, err := sql.Open("postgres", u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
		_, _ = admin.Exec("DROP DATABASE IF EXISTS " + name + " WITH (FORCE)")
		admin.Close()
	})
	if err := migrations.Up(db); err != nil {
		t.Fatalf("migrations: %v", err)
	}
	return db
}
