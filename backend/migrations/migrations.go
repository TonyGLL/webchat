// Package migrations embeds the versioned SQL migrations and applies them.
//
// Files follow golang-migrate naming: NNNNNN_description.up.sql / .down.sql.
// To change the schema add a new numbered pair; never edit an applied migration.
package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var files embed.FS

// baselineVersion is the version that matches databases created by hand from the old schema.sql.
const baselineVersion = 1

// Up applies all pending migrations. Databases created before migrations existed
// (tables present, no schema_migrations table) are adopted at the baseline version.
func Up(db *sql.DB) error {
	source, err := iofs.New(files, ".")
	if err != nil {
		return fmt.Errorf("open migration files: %w", err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("open migration driver: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	if _, _, err := m.Version(); errors.Is(err, migrate.ErrNilVersion) {
		var legacy bool
		if err := db.QueryRow(`SELECT to_regclass('public.users') IS NOT NULL`).Scan(&legacy); err != nil {
			return fmt.Errorf("inspect database: %w", err)
		}
		if legacy {
			slog.Info("adopting existing database at migration baseline", "version", baselineVersion)
			if err := m.Force(baselineVersion); err != nil {
				return fmt.Errorf("baseline existing database: %w", err)
			}
		}
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	version, _, _ := m.Version()
	slog.Info("database schema is up to date", "version", version)
	return nil
}
