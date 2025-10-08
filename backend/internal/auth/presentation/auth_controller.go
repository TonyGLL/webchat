package presentation

import (
	"errors"
	"net/http"

	"backend/internal/auth/application/dtos"
	"backend/internal/auth/application/usecases"
	shared_domain "backend/internal/shared/domain"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthController struct {
	LoginUseCase    *usecases.LoginUseCase
	RegisterUseCase *usecases.RegisterUseCase
	Validate        *validator.Validate
}

func NewAuthController(
	loginUseCase *usecases.LoginUseCase,
	registerUseCase *usecases.RegisterUseCase,
	validate *validator.Validate,
) *AuthController {
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

	response, err := ctrl.LoginUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		if errors.Is(err, shared_domain.ErrInvalidCredentials) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log in"})
		return
	}

	ctx.JSON(http.StatusOK, response)
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

	response, err := ctrl.RegisterUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		if errors.Is(err, shared_domain.ErrConflict) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "User with this email or username already exists"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	ctx.JSON(http.StatusCreated, response)
}
