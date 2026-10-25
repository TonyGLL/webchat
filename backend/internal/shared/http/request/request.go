package request

import (
	"backend/internal/shared/http"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// The bind errors are deliberately generic: validator and encoding/json messages name Go struct
// fields and types, which clients must not see. The detail is not logged either (it may echo input).
var (
	errInvalidBody   = errors.New("invalid request body")
	errInvalidParams = errors.New("invalid request parameters")
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
		return errInvalidBody
	}

	if err := validate.Struct(dto); err != nil {
		return errInvalidBody
	}

	return nil
}

func BindURI(ctx *gin.Context, dto interface{}, validate *validator.Validate) error {
	if err := ctx.ShouldBindUri(dto); err != nil {
		return errInvalidParams
	}

	if err := validate.Struct(dto); err != nil {
		return errInvalidBody
	}

	return nil
}
