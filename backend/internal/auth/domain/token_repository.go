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
}
