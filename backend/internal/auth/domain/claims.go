package domain

import "github.com/golang-jwt/jwt/v5"

// Token types carried in the "typ" claim.
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
	TokenTypeVerify  = "verify"
)

// CustomClaims represents the claims that will be encoded in the JWT.
// It embeds jwt.RegisteredClaims and adds custom fields.
type CustomClaims struct {
	UserID int `json:"user_id"`
	// Type says what the token may be used for. Every token is signed with the same secret, so
	// without it a refresh or e-mail verification token would pass as an access token.
	Type string `json:"typ"`
	jwt.RegisteredClaims
}

type EmailClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}
