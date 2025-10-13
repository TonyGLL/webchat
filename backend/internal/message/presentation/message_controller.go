package presentation

import (
	"backend/internal/message/application"
	"backend/internal/shared/http/request"
	"backend/internal/shared/http/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type MessageController struct {
	createMessageUseCase  *application.CreateMessageUseCase
	addReactionUseCase    *application.AddReactionUseCase
	removeReactionUseCase *application.RemoveReactionUseCase
	validate              *validator.Validate
}

func NewMessageController(createMessageUseCase *application.CreateMessageUseCase, addReactionUseCase *application.AddReactionUseCase, removeReactionUseCase *application.RemoveReactionUseCase, validate *validator.Validate) *MessageController {
	return &MessageController{createMessageUseCase: createMessageUseCase, addReactionUseCase: addReactionUseCase, removeReactionUseCase: removeReactionUseCase, validate: validate}
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
