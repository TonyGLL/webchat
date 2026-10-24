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
	auditor domain.Auditor,
) {
	createMessageUseCase := application.NewCreateMessageUseCase(messageRepository, wsBroadcaster)
	updateMessageUseCase := application.NewUpdateMessageUseCase(messageRepository, wsBroadcaster)
	addReactionUseCase := application.NewAddReactionUseCase(reactionRepository, messageRepository, wsBroadcaster)
	removeReactionUseCase := application.NewRemoveReactionUseCase(reactionRepository, messageRepository, wsBroadcaster)
	listMessagesUseCase := application.NewListMessagesUseCase(messageRepository)
	deleteMessageUseCase := application.NewDeleteMessageUseCase(messageRepository, wsBroadcaster, auditor)
	pinMessageUseCase := application.NewPinMessageUseCase(messageRepository, wsBroadcaster)
	listPinnedUseCase := application.NewListPinnedUseCase(messageRepository)
	searchMessagesUseCase := application.NewSearchMessagesUseCase(messageRepository)

	messageController := presentation.NewMessageController(
		createMessageUseCase,
		updateMessageUseCase,
		addReactionUseCase,
		removeReactionUseCase,
		listMessagesUseCase,
		deleteMessageUseCase,
		pinMessageUseCase,
		listPinnedUseCase,
		searchMessagesUseCase,
		validate,
	)

	// Register routes
	messageRoutes := router.Group("/messages")
	messageRoutes.Use(authMiddleware)
	{
		messageRoutes.GET("/search", messageController.SearchMessages)
		messageRoutes.GET("/:roomId/messages", messageController.FetchMessages)
		messageRoutes.POST("", messageController.CreateMessage)
		messageRoutes.PATCH("/:id", messageController.UpdateMessage)
		messageRoutes.DELETE("/:id", messageController.DeleteMessage)
		messageRoutes.PUT("/:id/pin", messageController.PinMessage)
		messageRoutes.DELETE("/:id/pin", messageController.UnpinMessage)
		messageRoutes.GET("/:roomId/pinned", messageController.FetchPinned)
		messageRoutes.POST("/:id/reactions", messageController.AddReaction)
		messageRoutes.DELETE("/:id/reactions", messageController.RemoveReaction)
	}
}
