package controllers

import (
	"backend/application/usecases"
	"backend/domain"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type MessageController struct {
	CreateMessageUseCase *usecases.CreateMessageUseCase
	Validate             *validator.Validate
}

func NewMessageController(createMessageUseCase *usecases.CreateMessageUseCase, validate *validator.Validate) *MessageController {
	return &MessageController{
		CreateMessageUseCase: createMessageUseCase,
		Validate:             validate,
	}
}

func (ctrl *MessageController) CreateMessage(c *gin.Context) {
	var input usecases.CreateMessageInputDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := ctrl.Validate.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message, err := ctrl.CreateMessageUseCase.Execute(input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, domain.ErrConflict):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
		}
		return
	}

	c.JSON(http.StatusCreated, message)
}