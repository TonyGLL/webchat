package auth

import (
	"backend/application/dtos"
	"backend/application/usecases"
	"backend/domain"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthController struct {
	LoginUseCase    *usecases.LoginUseCase
	RegisterUseCase *usecases.RegisterUseCase
	Validate        *validator.Validate
}

func NewAuthController(loginUseCase *usecases.LoginUseCase, registerUseCase *usecases.RegisterUseCase, validate *validator.Validate) *AuthController {
	return &AuthController{
		LoginUseCase:    loginUseCase,
		RegisterUseCase: registerUseCase,
		Validate:        validate,
	}
}

func (ctrl *AuthController) Login(ctx *gin.Context) {
	var input dtos.LoginInputDTO
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := ctrl.Validate.Struct(input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := ctrl.LoginUseCase.Execute(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, domain.ErrConflict):
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log in"})
		}
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (ctrl *AuthController) Register(ctx *gin.Context) {
	var input dtos.RegisterInputDTO
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := ctrl.Validate.Struct(input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := ctrl.RegisterUseCase.Execute(input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, domain.ErrConflict):
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		}
		return
	}

	ctx.JSON(http.StatusCreated, user)
}
