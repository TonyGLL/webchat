package application

import (
	"time"

	"backend/internal/auth/domain"
)

// JwtService defines the interface for JWT operations.
// This allows the application layer to be independent of the specific JWT implementation.
type JwtService interface {
	GenerateToken(userID int, duration time.Duration) (string, error)
	GenerateRefreshToken(userID int, tokenID string, duration time.Duration) (string, error)
	ParseToken(tokenString string) (*domain.CustomClaims, error)
}