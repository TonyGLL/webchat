package domain

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
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
	if _, reserved := reservedUsernames[normalizeForReserved(username)]; reserved {
		return ErrReservedUsername
	}
	return nil
}

// ErrReservedUsername is returned for names that would impersonate staff or the system.
var ErrReservedUsername = errors.New("this username is reserved, choose another one")

// ErrInvalidDisplayName is returned for empty names or names with control, bidi or invisible characters.
var ErrInvalidDisplayName = errors.New("name must not be empty or contain control or invisible characters")

// reservedUsernames are blocked at registration only (existing accounts keep theirs).
var reservedUsernames = func() map[string]struct{} {
	set := map[string]struct{}{}
	for _, n := range strings.Fields(`admin administrator administrador root superuser sudo moderator moderador mod mods
		support soporte system sistema staff help ayuda helpdesk security seguridad owner official oficial webchat
		bot api null undefined anonymous anonimo everyone here channel postmaster hostmaster webmaster abuse noreply
		no-reply mailer-daemon info contact team equipo`) {
		set[n] = struct{}{}
	}
	return set
}()

// normalizeForReserved folds case and the separators people use to dodge a blocklist
// ("Ad.min", "ad_min", "ADMIN").
func normalizeForReserved(username string) string {
	return strings.Map(func(r rune) rune {
		if r == '.' || r == '_' {
			return -1
		}
		return unicode.ToLower(r)
	}, username)
}

// NormalizeDisplayName returns the name trimmed, in Unicode NFC, with every run of whitespace
// collapsed to one space. Control characters, bidi overrides, zero-width and other invisible
// (format) characters are rejected, as is an empty result.
func NormalizeDisplayName(name string) (string, error) {
	name = norm.NFC.String(name)
	if !utf8.ValidString(name) {
		return "", ErrInvalidDisplayName
	}
	var b strings.Builder
	space := false
	for _, r := range name {
		switch {
		case unicode.IsSpace(r) && (r == ' ' || unicode.Is(unicode.Zs, r)):
			space = true
			continue
		case unicode.IsControl(r), unicode.Is(unicode.Cf, r), unicode.Is(unicode.Zl, r), unicode.Is(unicode.Zp, r),
			unicode.Is(unicode.Co, r), unicode.Is(unicode.Cs, r), r == unicode.ReplacementChar:
			return "", ErrInvalidDisplayName
		}
		if space && b.Len() > 0 {
			b.WriteByte(' ')
		}
		space = false
		b.WriteRune(r)
	}
	if b.Len() == 0 {
		return "", ErrInvalidDisplayName
	}
	return b.String(), nil
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
