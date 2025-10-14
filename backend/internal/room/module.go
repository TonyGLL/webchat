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
	router *gin.RouterGroup,
	validate *validator.Validate,
	authMiddleware gin.HandlerFunc,
	listUserRoomsUseCase *application.ListUserRoomsUseCase, // Pass the use case in
) {
	// Initialize other use cases
	createRoomUseCase := application.NewCreateRoomUseCase(roomRepo, memberRepo)
	joinRoomUseCase := application.NewJoinRoomUseCase(roomRepo, memberRepo)
	createInviteUseCase := application.NewCreateInviteUseCase(inviteRepo, roomRepo)
	acceptInviteUseCase := application.NewAcceptInviteUseCase(inviteRepo, memberRepo)

	// Initialize the controller, reusing the provided use case
	roomController := presentation.NewRoomController(
		createRoomUseCase,
		listUserRoomsUseCase,
		joinRoomUseCase,
		createInviteUseCase,
		acceptInviteUseCase,
		validate,
	)

	// Register routes under an authenticated group
	roomsRoutes := router.Group("/rooms")
	roomsRoutes.Use(authMiddleware)
	{
		roomsRoutes.POST("", roomController.CreateRoom)
		roomsRoutes.GET("", roomController.GetUserRooms)
		roomsRoutes.POST("/join/:room_id", roomController.JoinRoom)
		roomsRoutes.POST("/invite/:room_id", roomController.CreateInvite)
		roomsRoutes.POST("/accept-invite/:code", roomController.AcceptInvite)
	}
}
