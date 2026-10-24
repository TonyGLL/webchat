package room

import (
	"backend/internal/room/application"
	"backend/internal/room/domain"
	presentation "backend/internal/room/presentation"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// RegisterModule initializes the dependencies for the room module and registers its routes.
func RegisterModule(
	roomRepo domain.RoomRepository,
	memberRepo domain.RoomMemberRepository,
	inviteRepo domain.InviteRepository,
	invitationRepo domain.InvitationRepository,
	router *gin.RouterGroup,
	validate *validator.Validate,
	authMiddleware gin.HandlerFunc,
	listUserRoomsUseCase *application.ListUserRoomsUseCase, // Pass the use case in
	subscriber domain.RealtimeSubscriber,
	presence domain.PresenceChecker,
	bans domain.BanChecker,
) {
	// Initialize other use cases
	createRoomUseCase := application.NewCreateRoomUseCase(roomRepo, memberRepo, subscriber)
	joinRoomUseCase := application.NewJoinRoomUseCase(roomRepo, memberRepo, subscriber, bans)
	createInviteUseCase := application.NewCreateInviteUseCase(inviteRepo, roomRepo, memberRepo)
	acceptInviteUseCase := application.NewAcceptInviteUseCase(inviteRepo, memberRepo, subscriber, bans)

	markRoomReadUseCase := application.NewMarkRoomReadUseCase(memberRepo)
	leaveRoomUseCase := application.NewLeaveRoomUseCase(memberRepo, subscriber)
	listMembersUseCase := application.NewListMembersUseCase(memberRepo, presence)
	openDirectUseCase := application.NewOpenDirectUseCase(roomRepo, memberRepo, subscriber)
	discoverRoomsUseCase := application.NewDiscoverRoomsUseCase(roomRepo)
	inviteUserUseCase := application.NewInviteUserUseCase(invitationRepo, roomRepo, memberRepo, bans)
	listInvitationsUseCase := application.NewListInvitationsUseCase(invitationRepo)
	respondInvitationUseCase := application.NewRespondInvitationUseCase(invitationRepo, memberRepo, subscriber, bans)

	// Initialize the controller, reusing the provided use case
	roomController := presentation.NewRoomController(
		createRoomUseCase,
		listUserRoomsUseCase,
		markRoomReadUseCase,
		leaveRoomUseCase,
		listMembersUseCase,
		openDirectUseCase,
		joinRoomUseCase,
		createInviteUseCase,
		acceptInviteUseCase,
		discoverRoomsUseCase,
		inviteUserUseCase,
		listInvitationsUseCase,
		respondInvitationUseCase,
		validate,
	)

	// Register routes under an authenticated group
	roomsRoutes := router.Group("/rooms")
	roomsRoutes.Use(authMiddleware)
	{
		roomsRoutes.POST("", roomController.CreateRoom)
		roomsRoutes.GET("", roomController.GetUserRooms)
		roomsRoutes.GET("/discover", roomController.DiscoverRooms)
		roomsRoutes.POST("/invite-user/:room_id", roomController.InviteUser)
		roomsRoutes.GET("/invitations", roomController.ListInvitations)
		roomsRoutes.POST("/invitations/:id/accept", roomController.AcceptInvitation)
		roomsRoutes.POST("/invitations/:id/decline", roomController.DeclineInvitation)
		roomsRoutes.POST("/direct/:user_id", roomController.OpenDirect)
		roomsRoutes.GET("/members/:room_id", roomController.ListMembers)
		roomsRoutes.POST("/leave/:room_id", roomController.LeaveRoom)
		roomsRoutes.POST("/read/:room_id", roomController.MarkRead)
		roomsRoutes.POST("/join/:room_id", roomController.JoinRoom)
		roomsRoutes.POST("/invite/:room_id", roomController.CreateInvite)
		roomsRoutes.POST("/accept-invite/:code", roomController.AcceptInvite)
	}
}
