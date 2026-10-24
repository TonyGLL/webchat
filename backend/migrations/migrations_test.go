package migrations

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// TestUpOnEmptyDatabase applies every migration to a brand-new database, twice, as a new
// environment (CI, staging, a fresh cloud instance) would on its first start. It needs
// TEST_DATABASE_URL pointing at an EMPTY database and is skipped otherwise, so it can never
// touch a database that already holds data.
func TestUpOnEmptyDatabase(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	db, err := sql.Open("postgres", url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var used bool
	if err := db.QueryRow(`SELECT to_regclass('public.users') IS NOT NULL`).Scan(&used); err != nil {
		t.Fatalf("cannot reach the test database: %v", err)
	}
	if used {
		t.Skip("TEST_DATABASE_URL must point at an empty database")
	}

	if err := Up(db); err != nil {
		t.Fatalf("first Up: %v", err)
	}
	if err := Up(db); err != nil {
		t.Fatalf("Up must be idempotent: %v", err)
	}

	for _, table := range []string{"users", "rooms", "room_members", "messages", "reactions", "bans", "audit_log", "room_invitations"} {
		var exists bool
		if err := db.QueryRow(`SELECT to_regclass('public.' || $1) IS NOT NULL`, table).Scan(&exists); err != nil || !exists {
			t.Errorf("table %s missing after migrations (err=%v)", table, err)
		}
	}
}
