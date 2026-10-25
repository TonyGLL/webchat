package domain

import (
	_ "embed"
	"errors"
	"strings"
)

const (
	MinPasswordLength = 8
	// MaxPasswordLength is bcrypt's limit: longer passwords are rejected by the hasher.
	MaxPasswordLength = 72
)

var (
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong  = errors.New("password must be at most 72 characters")
	ErrPasswordCommon   = errors.New("password is too common, choose a less predictable one")
	ErrPasswordPersonal = errors.New("password must not contain your username or email")
)

// common_passwords.txt: lowercase passwords of 8+ characters taken from public breach corpora
// (SecLists 10k-most-common) plus Spanish and project-specific ones, one per line (~20 KB).
//
//go:embed common_passwords.txt
var commonPasswordsFile string

var commonPasswords = func() map[string]struct{} {
	set := map[string]struct{}{}
	for _, line := range strings.Split(commonPasswordsFile, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			set[line] = struct{}{}
		}
	}
	return set
}()

// ValidatePassword checks the password policy. username and email are the account's
// own identifiers, which must not appear inside the password.
func ValidatePassword(password, username, email string) error {
	switch {
	case len(password) < MinPasswordLength:
		return ErrPasswordTooShort
	case len(password) > MaxPasswordLength:
		return ErrPasswordTooLong
	}

	lower := strings.ToLower(password)
	if _, common := commonPasswords[lower]; common {
		return ErrPasswordCommon
	}

	if u := strings.ToLower(username); len(u) >= 3 && strings.Contains(lower, u) {
		return ErrPasswordPersonal
	}
	if local, _, found := strings.Cut(strings.ToLower(email), "@"); found && len(local) >= 3 && strings.Contains(lower, local) {
		return ErrPasswordPersonal
	}
	return nil
}
