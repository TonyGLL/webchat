package http

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	auth_domain "backend/internal/auth/domain"
	"backend/internal/shared/application"

	"github.com/gin-gonic/gin"
)

// CtxUserIDKey is the key for the user ID in the context.
const CtxUserIDKey = "userID"

// SessionGuard decides whether an otherwise valid access token still belongs to a live session:
// the account may have been deactivated or deleted, or its sessions revoked (password reset),
// after the token was issued.
type SessionGuard interface {
	Valid(ctx context.Context, userID int, issuedAt time.Time) bool
}

// JWTMiddleware creates a Gin middleware for JWT authentication. The guard may be nil (tests).
func JWTMiddleware(service application.JwtService, guard SessionGuard) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is not in 'Bearer {token}' format"})
			return
		}
		tokenStr := parts[1]

		claims, err := service.ParseToken(tokenStr)
		if err != nil {
			// Log the actual error for debugging, but return a generic error to the client.
			log.Printf("JWT parsing error: %v", err)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Only access tokens open the API: refresh and e-mail verification tokens share the
		// signing secret but must never authenticate a request. Tokens issued before the "typ"
		// claim existed are rejected too, so open sessions refresh once after the upgrade.
		if claims.Type != auth_domain.TokenTypeAccess || claims.ID != "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		if guard != nil {
			var issuedAt time.Time
			if claims.IssuedAt != nil {
				issuedAt = claims.IssuedAt.Time
			}
			if !guard.Valid(ctx.Request.Context(), claims.UserID, issuedAt) {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
				return
			}
		}

		// Set the user ID in the context for subsequent handlers.
		ctx.Set(CtxUserIDKey, claims.UserID)

		ctx.Next()
	}
}
