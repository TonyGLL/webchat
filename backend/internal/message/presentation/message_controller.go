package presentation

import (
	"backend/internal/message/application"
	"backend/internal/message/domain"
	shared_domain "backend/internal/shared/domain"
	"backend/internal/shared/http/request"
	"backend/internal/shared/http/response"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type MessageController struct {
	createMessageUseCase  *application.CreateMessageUseCase
	updateMessageUseCase  *application.UpdateMessageUseCase
	addReactionUseCase    *application.AddReactionUseCase
	removeReactionUseCase *application.RemoveReactionUseCase
	listMessagesUseCase   *application.ListMessagesUseCase
	validate              *validator.Validate
}

func NewMessageController(
	createMessageUseCase *application.CreateMessageUseCase,
	updateMessageUseCase *application.UpdateMessageUseCase,
	addReactionUseCase *application.AddReactionUseCase,
	removeReactionUseCase *application.RemoveReactionUseCase,
	listMessagesUseCase *application.ListMessagesUseCase,
	validate *validator.Validate,
) *MessageController {
	return &MessageController{
		createMessageUseCase:  createMessageUseCase,
		updateMessageUseCase:  updateMessageUseCase,
		addReactionUseCase:    addReactionUseCase,
		removeReactionUseCase: removeReactionUseCase,
		listMessagesUseCase:   listMessagesUseCase,
		validate:              validate,
	}
}

func (ctrl *MessageController) FetchMessages(c *gin.Context) {
	userID, err := request.GetUserID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}
	roomID := c.Param("roomId")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))

	messages, err := ctrl.listMessagesUseCase.Execute(c.Request.Context(), roomID, userID, page, pageSize)
	if err != nil {
		respondError(c, err, "Failed to fetch messages")
		return
	}

	response.JSON(c, http.StatusOK, messages)
}

func (ctrl *MessageController) CreateMessage(c *gin.Context) {
	userID, err := request.GetUserID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}
	var dto application.CreateMessageDTO
	if err := request.BindJSON(c, &dto, ctrl.validate); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	message, err := ctrl.createMessageUseCase.Execute(c.Request.Context(), dto, userID)
	if err != nil {
		respondError(c, err, "Failed to create message")
		return
	}
	response.JSON(c, http.StatusCreated, message)
}

func (ctrl *MessageController) UpdateMessage(ctx *gin.Context) {
	userID, err := request.GetUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}
	messageID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid message ID format")
		return
	}

	var dto application.UpdateMessageDTO
	if err := request.BindJSON(ctx, &dto, ctrl.validate); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := ctrl.updateMessageUseCase.Execute(ctx.Request.Context(), messageID, userID, dto); err != nil {
		respondError(ctx, err, "Failed to update message")
		return
	}

	response.JSON(ctx, http.StatusAccepted, gin.H{"message": "Message updated"})
}

func (ctrl *MessageController) AddReaction(c *gin.Context) {
	userID, err := request.GetUserID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}
	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid message ID format")
		return
	}
	var dto application.AddReactionDTO
	if err := request.BindJSON(c, &dto, ctrl.validate); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := ctrl.addReactionUseCase.Execute(c.Request.Context(), messageID, userID, dto); err != nil {
		respondError(c, err, "Failed to add reaction")
		return
	}
	response.JSON(c, http.StatusCreated, gin.H{"message": "Reaction added"})
}

func (ctrl *MessageController) RemoveReaction(c *gin.Context) {
	userID, err := request.GetUserID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}
	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid message ID format")
		return
	}
	var dto application.RemoveReactionDTO
	if err := request.BindJSON(c, &dto, ctrl.validate); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := ctrl.removeReactionUseCase.Execute(c.Request.Context(), messageID, userID, dto); err != nil {
		respondError(c, err, "Failed to remove reaction")
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"message": "Reaction removed"})
}

// respondError maps domain errors to HTTP statuses without leaking internal details.
func respondError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, shared_domain.ErrForbidden):
		response.Error(c, http.StatusForbidden, "You are not allowed to do this")
	case errors.Is(err, domain.ErrMessageNotFound):
		response.Error(c, http.StatusNotFound, "Message not found")
	case errors.Is(err, shared_domain.ErrInvalidInput):
		response.Error(c, http.StatusBadRequest, "Invalid input")
	default:
		log.Printf("%s: %v", fallback, err)
		response.Error(c, http.StatusInternalServerError, fallback)
	}
}
