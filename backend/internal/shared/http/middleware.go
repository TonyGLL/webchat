package http

import (
	"log"
	"net/http"
	"strings"

	"backend/internal/shared/application"

	"github.com/gin-gonic/gin"
)

// CtxUserIDKey is the key for the user ID in the context.
const CtxUserIDKey = "userID"

// JWTMiddleware creates a Gin middleware for JWT authentication.
func JWTMiddleware(service application.JwtService) gin.HandlerFunc {
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

		// Set the user ID in the context for subsequent handlers.
		ctx.Set(CtxUserIDKey, claims.UserID)

		ctx.Next()
	}
}
