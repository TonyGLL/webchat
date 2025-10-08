package message

import (
	"backend/internal/message/application"
	"backend/internal/message/domain"
	"backend/internal/message/presentation"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// RegisterModule initializes the dependencies for the message module and registers its routes.
func RegisterModule(
	messageRepository domain.MessageRepository,
	router *gin.RouterGroup,
	validate *validator.Validate,
	authMiddleware gin.HandlerFunc,
) {
	// Initialize use cases
	createMessageUseCase := application.NewCreateMessageUseCase(messageRepository)

	// Initialize the controller
	messageController := presentation.NewMessageController(createMessageUseCase, validate)

	// Register routes under an authenticated group
	messageRoutes := router.Group("/messages")
	messageRoutes.Use(authMiddleware) // Apply JWT middleware to all message routes
	{
		messageRoutes.POST("/", messageController.CreateMessage)
	}
}
