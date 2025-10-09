package auth

import (
	"backend/internal/auth/application/usecases"
	"backend/internal/auth/domain"
	"backend/internal/auth/presentation"
	shared_app "backend/internal/shared/application"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// RegisterModule initializes all the dependencies for the auth module and registers the routes.
// It acts as the dependency injection container for this module.
func RegisterModule(
	authRepository domain.AuthRepository,
	tokenRepository domain.TokenRepository,
	router *gin.RouterGroup,
	validate *validator.Validate,
	jwtService shared_app.JwtService,
	mailerService shared_app.MailerService,
	store shared_app.Store,
) {
	// Initialize services
	passwordService := domain.NewPasswordService()

	// Initialize use cases
	loginUseCase := usecases.NewLoginUseCase(authRepository, tokenRepository, passwordService, jwtService)
	registerUseCase := usecases.NewRegisterUseCase(authRepository, passwordService, jwtService, store)
	sendVerifyEmailUseCase := usecases.NewSendVerifyEmailUseCase(authRepository, jwtService, mailerService, store)
	refreshTokenUseCase := usecases.NewRefreshTokenUseCase(authRepository, tokenRepository, jwtService)

	// Initialize the controller
	authController := presentation.NewAuthController(loginUseCase, registerUseCase, refreshTokenUseCase, sendVerifyEmailUseCase, validate)

	// Register routes
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/login", authController.Login)
		authRoutes.POST("/register", authController.Register)
		authRoutes.POST("/refresh-token", authController.RefreshToken)
		authRoutes.POST("/send-verify-email", authController.SendVerifyEmail)
	}
}
