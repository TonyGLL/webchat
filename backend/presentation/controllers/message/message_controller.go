package message

import (
	"backend/application/dtos"
	"backend/application/usecases"
	"backend/presentation/http/middleware"
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
	// The Author ID is now retrieved from the context, which is set by the JWT middleware.
	// This is a critical security improvement.
	authorID, exists := c.Get(string(middleware.CtxUserIDKey))
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var input dtos.CreateMessageDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := ctrl.Validate.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// The controller now passes the authorID from the context to the use case.
	message, err := ctrl.CreateMessageUseCase.Execute(c.Request.Context(), input, authorID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
		return
	}

	c.JSON(http.StatusCreated, message)
}