package presentation

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"backend/internal/room/application"
	"backend/internal/room/domain"
	shared_domain "backend/internal/shared/domain"
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
	markRoomReadUseCase  *application.MarkRoomReadUseCase
	leaveRoomUseCase     *application.LeaveRoomUseCase
	listMembersUseCase   *application.ListMembersUseCase
	openDirectUseCase    *application.OpenDirectUseCase
	joinRoomUseCase      *application.JoinRoomUseCase
	createInviteUseCase  *application.CreateInviteUseCase
	acceptInviteUseCase  *application.AcceptInviteUseCase
	validate             *validator.Validate
}

// NewRoomController creates a new instance of RoomController.
func NewRoomController(
	createRoomUseCase *application.CreateRoomUseCase,
	listUserRoomsUseCase *application.ListUserRoomsUseCase,
	markRoomReadUseCase *application.MarkRoomReadUseCase,
	leaveRoomUseCase *application.LeaveRoomUseCase,
	listMembersUseCase *application.ListMembersUseCase,
	openDirectUseCase *application.OpenDirectUseCase,
	joinRoomUseCase *application.JoinRoomUseCase,
	createInviteUseCase *application.CreateInviteUseCase,
	acceptInviteUseCase *application.AcceptInviteUseCase,
	validate *validator.Validate,
) *RoomController {
	return &RoomController{
		createRoomUseCase:    createRoomUseCase,
		listUserRoomsUseCase: listUserRoomsUseCase,
		markRoomReadUseCase:  markRoomReadUseCase,
		leaveRoomUseCase:     leaveRoomUseCase,
		listMembersUseCase:   listMembersUseCase,
		openDirectUseCase:    openDirectUseCase,
		joinRoomUseCase:      joinRoomUseCase,
		createInviteUseCase:  createInviteUseCase,
		acceptInviteUseCase:  acceptInviteUseCase,
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
		switch {
		case errors.Is(err, shared_domain.ErrForbidden):
			response.Error(ctx, http.StatusForbidden, "Only the room owner or an admin can invite")
		case errors.Is(err, shared_domain.ErrInvalidInput):
			response.Error(ctx, http.StatusBadRequest, "Invalid input")
		default:
			log.Printf("failed to create invite: %v", err)
			response.Error(ctx, http.StatusInternalServerError, "Failed to create invite")
		}
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

// MarkRead handles the endpoint for POST /api/v1/rooms/read/:room_id
func (c *RoomController) MarkRead(ctx *gin.Context) {
	userID, err := request.GetUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}
	roomID, err := uuid.Parse(ctx.Param("room_id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid room ID format")
		return
	}
	if err := c.markRoomReadUseCase.Execute(ctx, roomID, userID); err != nil {
		if errors.Is(err, shared_domain.ErrForbidden) {
			response.Error(ctx, http.StatusForbidden, "You are not a member of this room")
			return
		}
		log.Printf("failed to mark room as read: %v", err)
		response.Error(ctx, http.StatusInternalServerError, "Failed to mark room as read")
		return
	}
	ctx.Status(http.StatusNoContent)
}

// OpenDirect handles the endpoint for POST /api/v1/rooms/direct/:user_id
func (c *RoomController) OpenDirect(ctx *gin.Context) {
	userID, err := request.GetUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}
	otherID, err := strconv.Atoi(ctx.Param("user_id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid user ID")
		return
	}
	room, err := c.openDirectUseCase.Execute(ctx, userID, otherID)
	switch {
	case err == nil:
		response.JSON(ctx, http.StatusOK, room)
	case errors.Is(err, shared_domain.ErrInvalidInput):
		response.Error(ctx, http.StatusBadRequest, "You cannot message yourself")
	case errors.Is(err, domain.ErrUserNotFound):
		response.Error(ctx, http.StatusNotFound, "User not found")
	default:
		log.Printf("failed to open direct conversation: %v", err)
		response.Error(ctx, http.StatusInternalServerError, "Failed to open conversation")
	}
}

// ListMembers handles the endpoint for GET /api/v1/rooms/members/:room_id
func (c *RoomController) ListMembers(ctx *gin.Context) {
	userID, err := request.GetUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}
	roomID, err := uuid.Parse(ctx.Param("room_id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid room ID format")
		return
	}
	members, err := c.listMembersUseCase.Execute(ctx, roomID, userID)
	if err != nil {
		if errors.Is(err, shared_domain.ErrForbidden) {
			response.Error(ctx, http.StatusForbidden, "You are not a member of this room")
			return
		}
		log.Printf("failed to list members: %v", err)
		response.Error(ctx, http.StatusInternalServerError, "Failed to list members")
		return
	}
	response.JSON(ctx, http.StatusOK, members)
}

// LeaveRoom handles the endpoint for POST /api/v1/rooms/leave/:room_id
func (c *RoomController) LeaveRoom(ctx *gin.Context) {
	userID, err := request.GetUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}
	roomID, err := uuid.Parse(ctx.Param("room_id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid room ID format")
		return
	}
	if err := c.leaveRoomUseCase.Execute(ctx, roomID, userID); err != nil {
		switch {
		case errors.Is(err, shared_domain.ErrForbidden):
			response.Error(ctx, http.StatusForbidden, "You are not a member of this room")
		case errors.Is(err, application.ErrOwnerCannotLeave):
			response.Error(ctx, http.StatusConflict, err.Error())
		default:
			log.Printf("failed to leave room: %v", err)
			response.Error(ctx, http.StatusInternalServerError, "Failed to leave room")
		}
		return
	}
	ctx.Status(http.StatusNoContent)
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
		case errors.Is(err, application.ErrRoomNotFound):
			response.Error(ctx, http.StatusNotFound, err.Error())
		default:
			log.Printf("failed to join room: %v", err)
			response.Error(ctx, http.StatusInternalServerError, "Failed to join room")
		}
		return
	}

	response.JSON(ctx, http.StatusOK, gin.H{"message": "Successfully joined room"})
}

func (c *RoomController) AcceptInvite(ctx *gin.Context) {
	userID, err := request.GetUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	var dto application.AcceptInviteDTO
	if err := request.BindURI(ctx, &dto, c.validate); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	invite, err := c.acceptInviteUseCase.Execute(ctx, dto, userID)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrAlreadyMember):
			response.Error(ctx, http.StatusConflict, err.Error())
		case errors.Is(err, domain.ErrInviteNotFound):
			response.Error(ctx, http.StatusNotFound, err.Error())
		default:
			log.Printf("failed to accept invite: %v", err)
			response.Error(ctx, http.StatusInternalServerError, "Failed to accept invite")
		}
		return
	}

	response.JSON(ctx, http.StatusOK, invite)
}
