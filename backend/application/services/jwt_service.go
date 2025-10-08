package services

import "time"

type Claims struct {
	UserID    int
	ExpiresAt time.Time
}

type JwtService interface {
	GenerateToken(userID int, ttl time.Duration) (string, error)
	ParseToken(token string) (*Claims, error)
}
