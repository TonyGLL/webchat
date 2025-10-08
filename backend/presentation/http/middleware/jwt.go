package middleware

import (
	"backend/application/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// keys usados en Gin Context
const (
	CtxUserIDKey = "userID"
	CtxRolesKey  = "roles"
)

// JWTMiddleware devuelve un middleware de gin que usa auth.Service
func JWTMiddleware(service services.JwtService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// leer header Authorization
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization header"})
			return
		}
		tokenStr := parts[1]

		// parsear token (interfaz application.Service)
		claims, err := service.ParseToken(tokenStr)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token: " + err.Error()})
			return
		}

		// inyectar en el gin.Context para handlers posteriores
		ctx.Set(CtxUserIDKey, claims.UserID)

		// continuar
		ctx.Next()
	}
}
