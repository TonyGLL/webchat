package users

import (
	auth_domain "backend/internal/auth/domain"
	shared_app "backend/internal/shared/application"
	"backend/internal/users/application/usecases"
	"backend/internal/users/domain"
	"backend/internal/users/presentation"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// RegisterModule initializes all the dependencies for the auth module and registers the routes.
// It acts as the dependency injection container for this module.
func RegisterModule(
	usersRepository domain.UsersRepository,
	router *gin.RouterGroup,
	validate *validator.Validate,
	mailerService shared_app.MailerService,
	store shared_app.Store,
	authMiddleware gin.HandlerFunc,
	realtime usecases.Realtime,
	presence usecases.PresenceChecker,
	sessions usecases.SessionRevoker,
) {
	// Initialize use cases
	deactivateUserUseCase := usecases.NewDeactivateUserUseCase(usersRepository, auth_domain.NewPasswordService(), sessions, realtime)
	getUserProfileUseCase := usecases.NewGetUserProfileUseCase(usersRepository)

	exportAccountUseCase := usecases.NewExportAccountUseCase(usersRepository)
	deleteAccountUseCase := usecases.NewDeleteAccountUseCase(usersRepository, auth_domain.NewPasswordService(), realtime)

	listDirectoryUseCase := usecases.NewListDirectoryUseCase(usersRepository, presence)

	// Initialize the controller
	usersController := presentation.NewUsersController(deactivateUserUseCase, getUserProfileUseCase, exportAccountUseCase, deleteAccountUseCase, listDirectoryUseCase, validate)

	// Register authenticated routes
	usersRoutes := router.Group("/users")
	usersRoutes.Use(authMiddleware)
	{
		usersRoutes.GET("/profile", usersController.GetUserProfile)
		usersRoutes.GET("/directory", usersController.ListDirectory)
		usersRoutes.DELETE("/deactivate", usersController.DeactivateUser)
		usersRoutes.GET("/export", usersController.ExportAccount)
		usersRoutes.DELETE("/me", usersController.DeleteAccount)
	}
}
