package services

import (
	"backend/domain"
	"time"
)

// JwtService defines the interface for JWT operations.
// This allows the application layer to be independent of the specific JWT implementation.
type JwtService interface {
	GenerateToken(userID int, duration time.Duration) (string, error)
	ParseToken(tokenString string) (*domain.CustomClaims, error)
}