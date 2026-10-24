// Command admin manages application-wide roles. Global roles are never granted through the
// HTTP API; whoever can run this command (with access to the database) decides who is an admin.
//
//	go run ./cmd/admin list
//	go run ./cmd/admin grant --email someone@example.com --role admin      (admin | moderator)
//	go run ./cmd/admin revoke --email someone@example.com
//
// It reads the same configuration as the server (CONFIG_FILE or environment variables).
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"backend/internal/shared/config"
	"backend/internal/shared/infra/db"
	"backend/internal/shared/permissions"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: admin list | grant --email E --role admin|moderator | revoke --email E")
	}
	command := args[0]

	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	email := flags.String("email", "", "email of the user")
	role := flags.String("role", "", "global role: admin or moderator")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	cfg, err := config.NewConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	database, err := db.NewDBPool(cfg.DatabaseURL, db.PoolSettings{
		MaxOpenConns:    cfg.DBMaxOpenConns,
		MaxIdleConns:    cfg.DBMaxIdleConns,
		ConnMaxLifetime: cfg.DBConnMaxLifetime,
		ConnMaxIdleTime: cfg.DBConnMaxIdleTime,
	})
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer database.Close()
	ctx := context.Background()

	switch command {
	case "list":
		return list(ctx, database)
	case "grant":
		if *email == "" || (*role != string(permissions.GlobalAdmin) && *role != string(permissions.GlobalModerator)) {
			return errors.New("grant needs --email and --role admin|moderator")
		}
		return setRole(ctx, database, *email, *role)
	case "revoke":
		if *email == "" {
			return errors.New("revoke needs --email")
		}
		return setRole(ctx, database, *email, string(permissions.GlobalUser))
	}
	return fmt.Errorf("unknown command %q", command)
}

func setRole(ctx context.Context, database *sql.DB, email, role string) error {
	res, err := database.ExecContext(ctx, `UPDATE users SET role = $1 WHERE lower(email) = lower($2) AND deleted = FALSE`, role, strings.TrimSpace(email))
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("no active user with email %q", email)
	}
	fmt.Printf("%s is now %q\n", email, role)
	return nil
}

func list(ctx context.Context, database *sql.DB) error {
	rows, err := database.QueryContext(ctx, `SELECT id, username, email, role FROM users WHERE role <> 'user' AND deleted = FALSE ORDER BY role, id`)
	if err != nil {
		return err
	}
	defer rows.Close()
	found := false
	for rows.Next() {
		var id int
		var username, email, role string
		if err := rows.Scan(&id, &username, &email, &role); err != nil {
			return err
		}
		fmt.Printf("%-10s #%d %s <%s>\n", role, id, username, email)
		found = true
	}
	if !found {
		fmt.Println("no users with a global role")
	}
	return rows.Err()
}
