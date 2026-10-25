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
	// Family ties every refresh token of one login together ("fam"): rotation hands the family
	// on, and presenting an already-rotated token revokes it. Empty on tokens issued before
	// families existed (legacy: the first rotation starts the family).
	Family string `json:"fam,omitempty"`
	// SessionStart is when the login happened, Unix seconds ("sat"); it never slides, so it
	// bounds the whole session. Zero on legacy tokens (their issue time stands in).
	SessionStart int64 `json:"sat,omitempty"`
	jwt.RegisteredClaims
}

type EmailClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}
