package presentation

import (
	"backend/internal/message/application"
	"backend/internal/shared/http/request"
	"backend/internal/shared/http/response"
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
	roomID := c.Param("roomId")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))

	messages, err := ctrl.listMessagesUseCase.Execute(c.Request.Context(), roomID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch messages")
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
		response.Error(c, http.StatusInternalServerError, "Failed to create message")
		return
	}
	response.JSON(c, http.StatusCreated, message)
}

func (ctrl *MessageController) UpdateMessage(ctx *gin.Context) {
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

	if err := ctrl.updateMessageUseCase.Execute(ctx.Request.Context(), messageID, dto); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
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
		response.Error(c, http.StatusInternalServerError, err.Error())
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
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"message": "Reaction removed"})
}
