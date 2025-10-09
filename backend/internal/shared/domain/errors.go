package domain

import (
	"errors"

	"github.com/gin-gonic/gin"
)

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrNotFound           = errors.New("resource not found")
	ErrConflict           = errors.New("resource conflict")
	ErrInternal           = errors.New("internal server error")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

func ErrorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
