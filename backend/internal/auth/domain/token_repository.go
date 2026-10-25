package domain

import (
	"context"
	"time"
)

// TokenRepository defines the interface for storing and retrieving tokens.
// This is used to manage refresh tokens in a persistent store like Redis.
type TokenRepository interface {
	StoreToken(ctx context.Context, userID int, tokenID string, key string, expiresIn time.Duration) error
	// GetUserIDByRefreshToken retrieves the user ID associated with a given refresh token.
	// If the token is not found, it should return a domain-specific error (e.g., ErrNotFound).
	GetToken(ctx context.Context, key string, tokenID string) (int, error)
	// DeleteRefreshToken deletes a specific refresh token from the store.
	DeleteToken(ctx context.Context, key string, tokenID string) error
	// ConsumeToken atomically reads and deletes a token (single use): of two concurrent callers
	// only one gets the user id, the other gets ErrNotFound.
	ConsumeToken(ctx context.Context, key string, tokenID string) (int, error)
	// StoreExclusiveToken stores a token as the user's only outstanding one for that key: the
	// previous one (if any) stops working.
	StoreExclusiveToken(ctx context.Context, userID int, tokenID string, key string, expiresIn time.Duration) error
	// RotateRefreshToken atomically exchanges a refresh token for its successor (see
	// RotateRequest) and reports what happened; the string is the successor id on a replay.
	RotateRefreshToken(ctx context.Context, req RotateRequest) (RotateOutcome, string, error)
	// RevokeUserSessions ends every session the user has open right now: refresh tokens issued
	// before this moment stop working. The marker only needs to outlive the longest refresh token.
	RevokeUserSessions(ctx context.Context, userID int, ttl time.Duration) error
	// SessionsRevokedAt returns when the user's sessions were last revoked (zero if never).
	SessionsRevokedAt(ctx context.Context, userID int) (time.Time, error)
}

// RotateOutcome is the result of a refresh-token rotation.
type RotateOutcome int

const (
	// RotateInvalid: unknown, expired, logged-out or revoked token.
	RotateInvalid RotateOutcome = iota
	// RotateOK: the token was exchanged; NewID is now the family's current token.
	RotateOK
	// RotateReplay: the token had already been exchanged a moment ago (inside the grace); the
	// successor returned is the one issued the first time, so no second chain starts.
	RotateReplay
	// RotateReuse: an already-exchanged token came back after the grace, which means a copy of it
	// exists somewhere else: the whole family has been revoked.
	RotateReuse
)

// RotateRequest describes one rotation.
type RotateRequest struct {
	UserID      int
	Family      string
	PresentedID string
	NewID       string
	// Legacy marks a token issued before families existed (no family claim): its first rotation
	// starts the family, named after the token's own id.
	Legacy bool
	// Grace is how long the exchanged token keeps answering with the same successor.
	Grace time.Duration
	// TTL is the lifetime of the successor (and of the reuse-detection marker).
	TTL time.Duration
}
