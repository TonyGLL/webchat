package config

import (
	"strings"
	"testing"
	"time"
)

func TestNewConfigJWTSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/db")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")

	tests := []struct {
		name    string
		secret  string
		wantErr bool
	}{
		{"missing", "", true},
		{"too short", "short", true},
		{"old default", "default-secret", true},
		{"long enough", strings.Repeat("a", 32), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("JWT_SECRET", tt.secret)
			_, err := NewConfig("")
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewConfigPool(t *testing.T) {
	base := func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/db")
		t.Setenv("REDIS_URL", "redis://localhost:6379/0")
		t.Setenv("JWT_SECRET", strings.Repeat("a", 32))
	}

	t.Run("defaults", func(t *testing.T) {
		base(t)
		cfg, err := NewConfig("")
		if err != nil {
			t.Fatal(err)
		}
		if cfg.DBMaxOpenConns != 25 || cfg.DBMaxIdleConns != 25 ||
			cfg.DBConnMaxLifetime != 30*time.Minute || cfg.DBConnMaxIdleTime != 5*time.Minute {
			t.Fatalf("unexpected defaults: %+v", cfg)
		}
	})

	t.Run("custom", func(t *testing.T) {
		base(t)
		t.Setenv("DB_MAX_OPEN_CONNS", "10")
		t.Setenv("DB_MAX_IDLE_CONNS", "5")
		t.Setenv("DB_CONN_MAX_LIFETIME", "1h")
		t.Setenv("DB_CONN_MAX_IDLE_TIME", "90s")
		cfg, err := NewConfig("")
		if err != nil {
			t.Fatal(err)
		}
		if cfg.DBMaxOpenConns != 10 || cfg.DBMaxIdleConns != 5 ||
			cfg.DBConnMaxLifetime != time.Hour || cfg.DBConnMaxIdleTime != 90*time.Second {
			t.Fatalf("unexpected values: %+v", cfg)
		}
	})

	t.Run("idle default follows a small open limit", func(t *testing.T) {
		base(t)
		t.Setenv("DB_MAX_OPEN_CONNS", "10")
		cfg, err := NewConfig("")
		if err != nil || cfg.DBMaxIdleConns != 10 {
			t.Fatalf("idle = %v, err = %v", cfg.DBMaxIdleConns, err)
		}
	})

	invalid := map[string]map[string]string{
		"zero open":         {"DB_MAX_OPEN_CONNS": "0"},
		"negative idle":     {"DB_MAX_IDLE_CONNS": "-1"},
		"not a number":      {"DB_MAX_OPEN_CONNS": "abc"},
		"idle above open":   {"DB_MAX_OPEN_CONNS": "5", "DB_MAX_IDLE_CONNS": "6"},
		"bad lifetime":      {"DB_CONN_MAX_LIFETIME": "soon"},
		"zero idle time":    {"DB_CONN_MAX_IDLE_TIME": "0s"},
		"negative lifetime": {"DB_CONN_MAX_LIFETIME": "-5m"},
	}
	for name, env := range invalid {
		t.Run(name, func(t *testing.T) {
			base(t)
			for k, v := range env {
				t.Setenv(k, v)
			}
			if _, err := NewConfig(""); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
