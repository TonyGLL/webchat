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
	frontendURL string,
	secureCookies bool,
) {
	// Initialize services
	passwordService := domain.NewPasswordService()

	// Initialize use cases
	loginUseCase := usecases.NewLoginUseCase(authRepository, tokenRepository, passwordService, jwtService)
	registerUseCase := usecases.NewRegisterUseCase(authRepository, passwordService, jwtService)
	sendVerifyEmailUseCase := usecases.NewSendVerifyEmailUseCase(authRepository, tokenRepository, jwtService, mailerService, store, frontendURL)
	refreshTokenUseCase := usecases.NewRefreshTokenUseCase(authRepository, tokenRepository, jwtService)
	verifyEmailUseCase := usecases.NewVerifyEmailUseCase(authRepository, tokenRepository, jwtService, mailerService)

	logoutUseCase := usecases.NewLogoutUseCase(tokenRepository, jwtService)
	requestPasswordResetUseCase := usecases.NewRequestPasswordResetUseCase(authRepository, tokenRepository, mailerService, frontendURL)
	resetPasswordUseCase := usecases.NewResetPasswordUseCase(authRepository, tokenRepository, passwordService)

	// Initialize the controller
	authController := presentation.NewAuthController(loginUseCase, registerUseCase, refreshTokenUseCase, sendVerifyEmailUseCase, verifyEmailUseCase, logoutUseCase, requestPasswordResetUseCase, resetPasswordUseCase, validate, secureCookies)

	// Register routes
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/login", authController.Login)
		authRoutes.POST("/register", authController.Register)
		authRoutes.POST("/refresh-token", authController.RefreshToken)
		authRoutes.POST("/logout", authController.Logout)
		authRoutes.POST("/forgot-password", authController.ForgotPassword)
		authRoutes.POST("/reset-password", authController.ResetPassword)
		authRoutes.POST("/send-verify-email/:user_id", authController.SendVerifyEmail)
		authRoutes.POST("/verify-email/:token", authController.VerifyEmail)
	}
}
