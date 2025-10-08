package domain

import "github.com/golang-jwt/jwt/v5"

// CustomClaims represents the claims that will be encoded in the JWT.
// It embeds jwt.RegisteredClaims and adds custom fields.
type CustomClaims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

type EmailClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}
