package auth

import (
	"backend/application/dtos"
	"backend/application/services"
	"backend/application/usecases"
	"backend/domain"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
)

type AuthController struct {
	LoginUseCase    *usecases.LoginUseCase
	RegisterUseCase *usecases.RegisterUseCase
	JwtService      services.JwtService
	Validate        *validator.Validate
}

func NewAuthController(
	loginUseCase *usecases.LoginUseCase,
	registerUseCase *usecases.RegisterUseCase,
	jwtService services.JwtService,
	validate *validator.Validate,
) *AuthController {
	return &AuthController{
		LoginUseCase:    loginUseCase,
		RegisterUseCase: registerUseCase,
		JwtService:      jwtService,
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

	// The controller now calls the use case with the request context.
	user, err := ctrl.LoginUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log in"})
		return
	}

	// Token generation is now a responsibility of the controller.
	token, err := ctrl.JwtService.GenerateToken(user.ID, 24*time.Hour)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	response := dtos.LoginResponseDTO{
		User:  *user,
		Token: token,
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

	// The controller calls the use case with the request context.
	newUser, err := ctrl.RegisterUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		var pgErr *pgconn.PgError
		// Check for a specific database error, like a unique constraint violation.
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // 23505 is unique_violation
			ctx.JSON(http.StatusConflict, gin.H{"error": "User with this email or username already exists"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	// Token generation is handled here after successful registration.
	token, err := ctrl.JwtService.GenerateToken(newUser.ID, 24*time.Hour)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	response := dtos.RegisterResponseDTO{
		ID:    newUser.ID,
		Token: token,
	}

	ctx.JSON(http.StatusCreated, response)
}
