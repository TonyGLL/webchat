package domain

import (
	"context"
	"time"
)

// TokenRepository defines the interface for storing and retrieving tokens.
// This is used to manage refresh tokens in a persistent store like Redis.
type TokenRepository interface {
	// StoreRefreshToken stores a refresh token for a given user with a specified expiry.
	StoreRefreshToken(ctx context.Context, userID int, tokenID string, expiresIn time.Duration) error
	// GetUserIDByRefreshToken retrieves the user ID associated with a given refresh token.
	// If the token is not found, it should return a domain-specific error (e.g., ErrNotFound).
	GetUserIDByRefreshToken(ctx context.Context, tokenID string) (int, error)
	// DeleteRefreshToken deletes a specific refresh token from the store.
	DeleteRefreshToken(ctx context.Context, tokenID string) error
}