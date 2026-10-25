package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name, password string
		want           error
	}{
		{"ok", "correct-horse-9", nil},
		{"too short", "abc123", ErrPasswordTooShort},
		{"too long", strings.Repeat("a", 73), ErrPasswordTooLong},
		{"common", "Password123", ErrPasswordCommon},
		{"common numeric", "12345678", ErrPasswordCommon},
		{"common from the embedded list", "Baseball", ErrPasswordCommon},
		{"common spanish", "Contraseña1", ErrPasswordCommon},
		{"contains username", "tonyRocks2026", ErrPasswordPersonal},
		{"contains email name", "xxMariaxx99", ErrPasswordPersonal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password, "tony", "maria@example.com")
			if !errors.Is(err, tt.want) {
				t.Fatalf("got %v, want %v", err, tt.want)
			}
		})
	}
}

func TestCommonPasswordListIsLoaded(t *testing.T) {
	if len(commonPasswords) < 2000 {
		t.Fatalf("blocklist has only %d entries", len(commonPasswords))
	}
	for p := range commonPasswords {
		if len(p) < MinPasswordLength || p != strings.ToLower(p) {
			t.Fatalf("entry %q must be lowercase and at least %d characters", p, MinPasswordLength)
		}
	}
}
