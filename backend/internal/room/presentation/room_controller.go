package presentation

import (
	"errors"
	"net/http"

	"backend/internal/room/application"
	"backend/internal/shared/http/request"
	"backend/internal/shared/http/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// RoomController handles HTTP requests related to rooms.
type RoomController struct {
	createRoomUseCase    *application.CreateRoomUseCase
	listUserRoomsUseCase *application.ListUserRoomsUseCase
	joinRoomUseCase      *application.JoinRoomUseCase
	createInviteUseCase  *application.CreateInviteUseCase
	validate             *validator.Validate
}

// NewRoomController creates a new instance of RoomController.
func NewRoomController(
	createRoomUseCase *application.CreateRoomUseCase,
	listUserRoomsUseCase *application.ListUserRoomsUseCase,
	joinRoomUseCase *application.JoinRoomUseCase,
	createInviteUseCase *application.CreateInviteUseCase,
	validate *validator.Validate,
) *RoomController {
	return &RoomController{
		createRoomUseCase:    createRoomUseCase,
		listUserRoomsUseCase: listUserRoomsUseCase,
		joinRoomUseCase:      joinRoomUseCase,
		createInviteUseCase:  createInviteUseCase,
		validate:             validate,
	}
}

func (c *RoomController) CreateInvite(ctx *gin.Context) {
	userID, err := request.GetUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	var dto application.CreateInviteDTO
	dto.RoomID = ctx.Param("room_id")
	if err := request.BindJSON(ctx, &dto, c.validate); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	invite, err := c.createInviteUseCase.Execute(ctx.Request.Context(), dto, userID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "Failed to create invite")
		return
	}

	response.JSON(ctx, http.StatusCreated, invite)
}

// CreateRoom handles the endpoint for POST /api/v1/rooms
func (c *RoomController) CreateRoom(ctx *gin.Context) {
	userID, err := request.GetUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	var dto application.CreateRoomDTO
	if err := request.BindJSON(ctx, &dto, c.validate); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	room, err := c.createRoomUseCase.Execute(ctx, dto, userID)
	if err != nil {
		// Here you could map domain errors to HTTP status codes
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(ctx, http.StatusCreated, room)
}

// GetUserRooms handles the endpoint for GET /api/v1/rooms
func (c *RoomController) GetUserRooms(ctx *gin.Context) {
	userID, err := request.GetUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	rooms, err := c.listUserRoomsUseCase.Execute(ctx, userID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(ctx, http.StatusOK, rooms)
}

// JoinRoom handles the endpoint for POST /api/v1/rooms/:id/join
func (c *RoomController) JoinRoom(ctx *gin.Context) {
	userID, err := request.GetUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	roomIDStr := ctx.Param("room_id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid room ID format")
		return
	}

	err = c.joinRoomUseCase.Execute(ctx, roomID, userID)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrRoomIsPrivate):
			response.Error(ctx, http.StatusForbidden, err.Error())
		case errors.Is(err, application.ErrAlreadyMember):
			response.Error(ctx, http.StatusConflict, err.Error())
		default:
			response.Error(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.JSON(ctx, http.StatusOK, gin.H{"message": "Successfully joined room"})
}
