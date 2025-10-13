package request

import (
	"backend/internal/shared/http"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// GetUserID retrieves the user ID from the Gin context.
func GetUserID(c *gin.Context) (int, error) {
	userIDVal, exists := c.Get(string(http.CtxUserIDKey))
	if !exists {
		return 0, errors.New("user not authenticated: user_id not found in context")
	}

	userID, ok := userIDVal.(int)
	if !ok {
		return 0, errors.New("invalid user_id type in context")
	}

	return userID, nil
}

// BindJSON binds the request body to a DTO and validates it.
func BindJSON(ctx *gin.Context, dto interface{}, validate *validator.Validate) error {
	if err := ctx.ShouldBindJSON(dto); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}

	if err := validate.Struct(dto); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

func BindURI(ctx *gin.Context, dto interface{}, validate *validator.Validate) error {
	if err := ctx.ShouldBindUri(dto); err != nil {
		return fmt.Errorf("invalid URI parameters: %w", err)
	}

	if err := validate.Struct(dto); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}
