package middleware

import (
	"backend/application/services"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// contextKey is a private type for context keys to prevent collisions.
type contextKey string

// CtxUserIDKey is the key for the user ID in the context.
const CtxUserIDKey = contextKey("userID")

// JWTMiddleware creates a Gin middleware for JWT authentication.
func JWTMiddleware(service services.JwtService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is not in 'Bearer {token}' format"})
			return
		}
		tokenStr := parts[1]

		claims, err := service.ParseToken(tokenStr)
		if err != nil {
			// Log the actual error for debugging, but return a generic error to the client.
			log.Printf("JWT parsing error: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Set the user ID in the context for subsequent handlers.
		c.Set(string(CtxUserIDKey), claims.UserID)

		c.Next()
	}
}
