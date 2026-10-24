package domain

import (
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

var commonPasswords = map[string]struct{}{
	"12345678": {}, "123456789": {}, "1234567890": {}, "password": {}, "password1": {},
	"password123": {}, "qwertyui": {}, "qwerty123": {}, "11111111": {}, "00000000": {},
	"abc12345": {}, "iloveyou": {}, "letmein1": {}, "admin123": {}, "welcome1": {},
	"contraseña": {}, "contrasena": {}, "12341234": {}, "87654321": {}, "asdfghjk": {},
}

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
