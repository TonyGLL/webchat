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
	createRoomUseCase        *application.CreateRoomUseCase
	listUserRoomsUseCase     *application.ListUserRoomsUseCase
	markRoomReadUseCase      *application.MarkRoomReadUseCase
	leaveRoomUseCase         *application.LeaveRoomUseCase
	listMembersUseCase       *application.ListMembersUseCase
	openDirectUseCase        *application.OpenDirectUseCase
	joinRoomUseCase          *application.JoinRoomUseCase
	createInviteUseCase      *application.CreateInviteUseCase
	acceptInviteUseCase      *application.AcceptInviteUseCase
	discoverRoomsUseCase     *application.DiscoverRoomsUseCase
	inviteUserUseCase        *application.InviteUserUseCase
	listInvitationsUseCase   *application.ListInvitationsUseCase
	respondInvitationUseCase *application.RespondInvitationUseCase
	validate                 *validator.Validate
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
	discoverRoomsUseCase *application.DiscoverRoomsUseCase,
	inviteUserUseCase *application.InviteUserUseCase,
	listInvitationsUseCase *application.ListInvitationsUseCase,
	respondInvitationUseCase *application.RespondInvitationUseCase,
	validate *validator.Validate,
) *RoomController {
	return &RoomController{
		createRoomUseCase:        createRoomUseCase,
		listUserRoomsUseCase:     listUserRoomsUseCase,
		markRoomReadUseCase:      markRoomReadUseCase,
		leaveRoomUseCase:         leaveRoomUseCase,
		listMembersUseCase:       listMembersUseCase,
		openDirectUseCase:        openDirectUseCase,
		joinRoomUseCase:          joinRoomUseCase,
		createInviteUseCase:      createInviteUseCase,
		acceptInviteUseCase:      acceptInviteUseCase,
		discoverRoomsUseCase:     discoverRoomsUseCase,
		inviteUserUseCase:        inviteUserUseCase,
		listInvitationsUseCase:   listInvitationsUseCase,
		respondInvitationUseCase: respondInvitationUseCase,
		validate:                 validate,
	}
}

// queryInt reads an optional non-negative integer query parameter (0 when absent). On a malformed
// value it answers 400 itself and reports false.
func queryInt(ctx *gin.Context, name string) (int, bool) {
	raw := ctx.Query(name)
	if raw == "" {
		return 0, true
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		response.Error(ctx, http.StatusBadRequest, "Invalid "+name)
		return 0, false
	}
	return n, true
}

func (c *RoomController) CreateInvite(ctx *gin.Context) {
	userID, err := request.GetUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "Unauthorized")
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
			response.Error(ctx, http.StatusBadRequest, "Invalid input: expires_in must be between 1m and 720h")
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
		response.Error(ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var dto application.CreateRoomDTO
	if err := request.BindJSON(ctx, &dto, c.validate); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	room, err := c.createRoomUseCase.Execute(ctx, dto, userID)
	if err != nil {
		log.Printf("failed to create room: %v", err)
		response.Error(ctx, http.StatusInternalServerError, "Failed to create room")
		return
	}

	response.JSON(ctx, http.StatusCreated, room)
}

// GetUserRooms handles the endpoint for GET /api/v1/rooms
func (c *RoomController) GetUserRooms(ctx *gin.Context) {
	userID, err := request.GetUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}

	rooms, err := c.listUserRoomsUseCase.Execute(ctx, userID)
	if err != nil {
		log.Printf("failed to list rooms: %v", err)
		response.Error(ctx, http.StatusInternalServerError, "Failed to list rooms")
		return
	}

	response.JSON(ctx, http.StatusOK, rooms)
}

// MarkRead handles the endpoint for POST /api/v1/rooms/read/:room_id
func (c *RoomController) MarkRead(ctx *gin.Context) {
	userID, err := request.GetUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "Unauthorized")
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
		response.Error(ctx, http.StatusUnauthorized, "Unauthorized")
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
		response.Error(ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}
	roomID, err := uuid.Parse(ctx.Param("room_id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid room ID format")
		return
	}
	limit, ok := queryInt(ctx, "limit")
	if !ok {
		return
	}
	offset, ok := queryInt(ctx, "offset")
	if !ok {
		return
	}
	members, err := c.listMembersUseCase.Execute(ctx, roomID, userID, limit, offset)
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
		response.Error(ctx, http.StatusUnauthorized, "Unauthorized")
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
		response.Error(ctx, http.StatusUnauthorized, "Unauthorized")
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
		case errors.Is(err, application.ErrRoomIsPrivate), errors.Is(err, application.ErrBannedFromRoom):
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
		response.Error(ctx, http.StatusUnauthorized, "Unauthorized")
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
		case errors.Is(err, application.ErrBannedFromRoom):
			response.Error(ctx, http.StatusForbidden, err.Error())
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

// DiscoverRooms handles the endpoint for GET /api/v1/rooms/discover
func (c *RoomController) DiscoverRooms(ctx *gin.Context) {
	userID, err := request.GetUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limit, ok := queryInt(ctx, "limit")
	if !ok {
		return
	}
	var after *uuid.UUID
	if raw := ctx.Query("after"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			response.Error(ctx, http.StatusBadRequest, "Invalid after cursor")
			return
		}
		after = &id
	}

	rooms, err := c.discoverRoomsUseCase.Execute(ctx, userID, limit, after)
	if err != nil {
		log.Printf("failed to discover rooms: %v", err)
		response.Error(ctx, http.StatusInternalServerError, "Failed to discover rooms")
		return
	}

	response.JSON(ctx, http.StatusOK, rooms)
}

// InviteUser handles the endpoint for POST /api/v1/rooms/invite-user/:room_id
func (c *RoomController) InviteUser(ctx *gin.Context) {
	userID, err := request.GetUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}
	roomID, err := uuid.Parse(ctx.Param("room_id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid room ID format")
		return
	}
	var body struct {
		UserID int `json:"user_id" validate:"required,gt=0"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil || c.validate.Struct(body) != nil {
		response.Error(ctx, http.StatusBadRequest, "user_id is required")
		return
	}

	invitation, err := c.inviteUserUseCase.Execute(ctx, roomID, userID, body.UserID)
	switch {
	case err == nil:
		response.JSON(ctx, http.StatusCreated, invitation)
	case errors.Is(err, shared_domain.ErrInvalidInput):
		response.Error(ctx, http.StatusBadRequest, "You cannot invite yourself")
	case errors.Is(err, shared_domain.ErrForbidden):
		response.Error(ctx, http.StatusForbidden, "You cannot invite people to this room")
	case errors.Is(err, application.ErrBannedFromRoom):
		response.Error(ctx, http.StatusForbidden, err.Error())
	case errors.Is(err, application.ErrRoomNotFound), errors.Is(err, domain.ErrUserNotFound):
		response.Error(ctx, http.StatusNotFound, "Room or user not found")
	case errors.Is(err, application.ErrAlreadyMember), errors.Is(err, domain.ErrAlreadyInvited):
		response.Error(ctx, http.StatusConflict, err.Error())
	default:
		log.Printf("failed to invite user: %v", err)
		response.Error(ctx, http.StatusInternalServerError, "Failed to invite user")
	}
}

// ListInvitations handles the endpoint for GET /api/v1/rooms/invitations
func (c *RoomController) ListInvitations(ctx *gin.Context) {
	userID, err := request.GetUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}
	invitations, err := c.listInvitationsUseCase.Execute(ctx, userID)
	if err != nil {
		log.Printf("failed to list invitations: %v", err)
		response.Error(ctx, http.StatusInternalServerError, "Failed to list invitations")
		return
	}
	response.JSON(ctx, http.StatusOK, invitations)
}

func (c *RoomController) respondInvitation(ctx *gin.Context, accept bool) {
	userID, err := request.GetUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "Invalid invitation ID format")
		return
	}
	err = c.respondInvitationUseCase.Execute(ctx, id, userID, accept)
	switch {
	case err == nil:
		response.JSON(ctx, http.StatusOK, gin.H{"message": "Invitation answered"})
	case errors.Is(err, shared_domain.ErrNotFound):
		response.Error(ctx, http.StatusNotFound, "Invitation not found")
	case errors.Is(err, application.ErrBannedFromRoom):
		response.Error(ctx, http.StatusForbidden, err.Error())
	default:
		log.Printf("failed to answer invitation: %v", err)
		response.Error(ctx, http.StatusInternalServerError, "Failed to answer invitation")
	}
}

// AcceptInvitation handles POST /api/v1/rooms/invitations/:id/accept
func (c *RoomController) AcceptInvitation(ctx *gin.Context) { c.respondInvitation(ctx, true) }

// DeclineInvitation handles POST /api/v1/rooms/invitations/:id/decline
func (c *RoomController) DeclineInvitation(ctx *gin.Context) { c.respondInvitation(ctx, false) }
