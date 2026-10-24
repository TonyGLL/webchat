package config

import (
	"strings"
	"testing"
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
