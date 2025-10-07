package controllers

import (
	"backend/application/usecases"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MessageController struct {
	CreateMessageUseCase *usecases.CreateMessageUseCase
}

func NewMessageController(createMessageUseCase *usecases.CreateMessageUseCase) *MessageController {
	return &MessageController{CreateMessageUseCase: createMessageUseCase}
}

func (ctrl *MessageController) CreateMessage(c *gin.Context) {
	var input usecases.CreateMessageInputDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message, err := ctrl.CreateMessageUseCase.Execute(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
		return
	}

	c.JSON(http.StatusCreated, message)
}