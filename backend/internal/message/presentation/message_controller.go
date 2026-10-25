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
	deleteMessageUseCase  *application.DeleteMessageUseCase
	pinMessageUseCase     *application.PinMessageUseCase
	listPinnedUseCase     *application.ListPinnedUseCase
	searchMessagesUseCase *application.SearchMessagesUseCase
	validate              *validator.Validate
}

func NewMessageController(
	createMessageUseCase *application.CreateMessageUseCase,
	updateMessageUseCase *application.UpdateMessageUseCase,
	addReactionUseCase *application.AddReactionUseCase,
	removeReactionUseCase *application.RemoveReactionUseCase,
	listMessagesUseCase *application.ListMessagesUseCase,
	deleteMessageUseCase *application.DeleteMessageUseCase,
	pinMessageUseCase *application.PinMessageUseCase,
	listPinnedUseCase *application.ListPinnedUseCase,
	searchMessagesUseCase *application.SearchMessagesUseCase,
	validate *validator.Validate,
) *MessageController {
	return &MessageController{
		createMessageUseCase:  createMessageUseCase,
		updateMessageUseCase:  updateMessageUseCase,
		addReactionUseCase:    addReactionUseCase,
		removeReactionUseCase: removeReactionUseCase,
		listMessagesUseCase:   listMessagesUseCase,
		deleteMessageUseCase:  deleteMessageUseCase,
		pinMessageUseCase:     pinMessageUseCase,
		listPinnedUseCase:     listPinnedUseCase,
		searchMessagesUseCase: searchMessagesUseCase,
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

	// Cursor pagination: ?limit=50&before=<id of the oldest message already loaded>
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	before := c.Query("before")

	messages, err := ctrl.listMessagesUseCase.Execute(c.Request.Context(), roomID, userID, limit, before)
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

func (ctrl *MessageController) DeleteMessage(c *gin.Context) {
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
	if err := ctrl.deleteMessageUseCase.Execute(c.Request.Context(), messageID, userID); err != nil {
		respondError(c, err, "Failed to delete message")
		return
	}
	c.Status(http.StatusNoContent)
}

func (ctrl *MessageController) setPinned(c *gin.Context, pinned bool) {
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
	if err := ctrl.pinMessageUseCase.Execute(c.Request.Context(), messageID, userID, pinned); err != nil {
		respondError(c, err, "Failed to update pin")
		return
	}
	c.Status(http.StatusNoContent)
}

func (ctrl *MessageController) PinMessage(c *gin.Context)   { ctrl.setPinned(c, true) }
func (ctrl *MessageController) UnpinMessage(c *gin.Context) { ctrl.setPinned(c, false) }

func (ctrl *MessageController) FetchPinned(c *gin.Context) {
	userID, err := request.GetUserID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}
	messages, err := ctrl.listPinnedUseCase.Execute(c.Request.Context(), c.Param("roomId"), userID)
	if err != nil {
		respondError(c, err, "Failed to fetch pinned messages")
		return
	}
	response.JSON(c, http.StatusOK, messages)
}

// SearchMessages handles GET /messages/search?q=text[&room_id=uuid][&limit=n]
func (ctrl *MessageController) SearchMessages(c *gin.Context) {
	userID, err := request.GetUserID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	messages, err := ctrl.searchMessagesUseCase.Execute(c.Request.Context(), userID, c.Query("q"), c.Query("room_id"), limit)
	if err != nil {
		respondError(c, err, "Failed to search messages")
		return
	}
	response.JSON(c, http.StatusOK, messages)
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
	case errors.Is(err, domain.ErrSearchBusy), errors.Is(err, domain.ErrSearchTooBroad):
		c.Header("Retry-After", "5")
		response.Error(c, http.StatusServiceUnavailable, "Search is busy or too broad, narrow your search and try again")
	default:
		log.Printf("%s: %v", fallback, err)
		response.Error(c, http.StatusInternalServerError, fallback)
	}
}
