package domain

import (
	"errors"
	"regexp"
	"strings"
)

// ErrInvalidUsername is returned when a username has characters outside the allowed set.
var ErrInvalidUsername = errors.New("username must be 3 to 30 characters: letters, digits, '_', '.' or '-'")

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_.-]{3,30}$`)

// ValidateUsername restricts usernames to a charset without '@', so a username can never be
// mistaken for (or squat on) someone's email address at login.
func ValidateUsername(username string) error {
	if !usernamePattern.MatchString(username) {
		return ErrInvalidUsername
	}
	return nil
}

// NormalizeEmail trims and lowercases an email: addresses are unique regardless of case.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// LooksLikeEmail tells which identity a login or reset input refers to: usernames never
// contain '@'.
func LooksLikeEmail(input string) bool {
	return strings.Contains(input, "@")
}
