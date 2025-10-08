package auth

import (
	"backend/internal/auth/application"
	"backend/internal/auth/domain"
	"backend/internal/auth/persistence"
	"backend/internal/auth/presentation"
	shared_app "backend/internal/shared/application"
	"backend/internal/shared/infra/db"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// RegisterModule initializes all the dependencies for the auth module and registers the routes.
// It acts as the dependency injection container for this module.
func RegisterModule(
	store *db.SQLStore,
	router *gin.RouterGroup,
	validate *validator.Validate,
	jwtService shared_app.JwtService,
) {
	// Initialize services and repositories
	passwordService := domain.NewPasswordService()
	authRepository := persistence.NewPgAuthRepository(store)

	// Initialize use cases
	loginUseCase := application.NewLoginUseCase(store, passwordService, jwtService)
	registerUseCase := application.NewRegisterUseCase(store, passwordService, jwtService)

	// Initialize the controller
	authController := presentation.NewAuthController(loginUseCase, registerUseCase, validate)

	// Register routes
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/login", authController.Login)
		authRoutes.POST("/register", authController.Register)
	}
}