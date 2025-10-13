package message

import (
	"backend/internal/message/application"
	"backend/internal/message/domain"
	"backend/internal/message/presentation"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func RegisterModule(
	messageRepository domain.MessageRepository,
	reactionRepository domain.ReactionRepository,
	wsBroadcaster domain.WebsocketBroadcaster,
	router *gin.RouterGroup,
	validate *validator.Validate,
	authMiddleware gin.HandlerFunc,
) {
	createMessageUseCase := application.NewCreateMessageUseCase(messageRepository, wsBroadcaster)
	addReactionUseCase := application.NewAddReactionUseCase(reactionRepository, messageRepository, wsBroadcaster)
	removeReactionUseCase := application.NewRemoveReactionUseCase(reactionRepository, messageRepository, wsBroadcaster)
	messageController := presentation.NewMessageController(createMessageUseCase, addReactionUseCase, removeReactionUseCase, validate)
	messageRoutes := router.Group("/messages")
	messageRoutes.Use(authMiddleware)
	{
		messageRoutes.POST("/", messageController.CreateMessage)
		messageRoutes.POST("/:id/reactions", messageController.AddReaction)
		messageRoutes.DELETE("/:id/reactions", messageController.RemoveReaction)
	}
}
