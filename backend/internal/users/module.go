package users

import (
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
) {
	// Initialize use cases
	deactivateUserUseCase := usecases.NewDeactivateUserUseCase(usersRepository)
	getUserProfileUseCase := usecases.NewGetUserProfileUseCase(usersRepository)

	// Initialize the controller
	usersController := presentation.NewUsersController(deactivateUserUseCase, getUserProfileUseCase, validate)

	// Register authenticated routes
	usersRoutes := router.Group("/users")
	usersRoutes.Use(authMiddleware)
	{
		usersRoutes.GET("/:id", usersController.GetUserProfile)
		usersRoutes.DELETE("/deactivate", usersController.DeactivateUser)
	}
}
