package jwt

import (
	"backend/application/services"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secret []byte
}

func NewJWTService() services.JwtService {
	secret := os.Getenv("JWT_SECRET")
	return &JWTService{secret: []byte(secret)}
}

type myClaims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func (s *JWTService) GenerateToken(userID int, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := myClaims{
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

func (s *JWTService) ParseToken(tokenStr string) (*services.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &myClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := token.Claims.(*myClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token invalid")
	}
	return &services.Claims{
		UserID:    c.UserID,
		ExpiresAt: c.ExpiresAt.Time,
	}, nil
}
