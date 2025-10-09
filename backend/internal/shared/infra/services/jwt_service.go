package services

import (
	"errors"
	"strconv"
	"time"

	"backend/internal/auth/domain"
	"backend/internal/shared/application"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService is the infrastructure-level implementation of the JwtService interface.
type JWTService struct {
	secret []byte
}

// NewJWTService creates a new JWTService with the provided secret.
func NewJWTService(secret string) (application.JwtService, error) {
	if secret == "" {
		return nil, errors.New("JWT secret cannot be empty")
	}
	return &JWTService{secret: []byte(secret)}, nil
}

func (s *JWTService) GenerateTokenFromClaims(claims map[string]interface{}, ttl time.Duration) (string, error) {
	now := time.Now()
	stdClaims := jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(ttl).Unix(),
	}

	for k, v := range claims {
		stdClaims[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, stdClaims)
	return token.SignedString([]byte(s.secret))
}

// GenerateToken creates a new JWT for a given user ID and duration.
func (s *JWTService) GenerateToken(userID int, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := &domain.CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Subject:   strconv.FormatInt(int64(userID), 10),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// GenerateRefreshToken creates a new refresh JWT for a given user ID and token ID.
func (s *JWTService) GenerateRefreshToken(userID int, tokenID string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := &domain.CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Subject:   strconv.FormatInt(int64(userID), 10),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ParseToken validates a token string and returns the custom claims if valid.
func (s *JWTService) ParseToken(tokenStr string) (*domain.CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &domain.CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*domain.CustomClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token invalid")
	}

	return claims, nil
}
