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

	checkIdentityMigrationRefusesDuplicates(t, db)
}

// checkIdentityMigrationRefusesDuplicates verifies that 000013 fails loudly, without changing
// data, when two accounts already collide under lower(). Everything runs in transactions that
// are rolled back.
func checkIdentityMigrationRefusesDuplicates(t *testing.T, db *sql.DB) {
	t.Helper()
	up, err := files.ReadFile("000013_auth_case_insensitive_identity.up.sql")
	if err != nil {
		t.Fatal(err)
	}

	run := func(users [][3]string) error {
		tx, err := db.Begin()
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback()
		for _, stmt := range []string{`DROP INDEX users_email_lower_key`, `DROP INDEX users_username_lower_key`} {
			if _, err := tx.Exec(stmt); err != nil {
				t.Fatal(err)
			}
		}
		for _, u := range users {
			if _, err := tx.Exec(`INSERT INTO users (name, username, email) VALUES ($1, $2, $3)`, u[0], u[1], u[2]); err != nil {
				t.Fatal(err)
			}
		}
		_, err = tx.Exec(string(up))
		return err
	}

	if err := run([][3]string{{"A", "ana", "ana@example.com"}, {"B", "Ana", "other@example.com"}}); err == nil {
		t.Error("migration must fail when usernames collide case-insensitively")
	}
	if err := run([][3]string{{"A", "ana", "ana@example.com"}, {"B", "bob", "ANA@example.com"}}); err == nil {
		t.Error("migration must fail when emails collide case-insensitively")
	}
	if err := run([][3]string{{"A", "ana", "ana@example.com"}, {"B", "bob", "bob@example.com"}}); err != nil {
		t.Errorf("migration must pass on clean data: %v", err)
	}
}
