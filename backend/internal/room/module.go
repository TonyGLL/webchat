package room

import (
	"backend/internal/room/application"
	"backend/internal/room/domain"
	presentation "backend/internal/room/presentation"
	shared_domain "backend/internal/shared/domain"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// RegisterModule initializes the dependencies for the room module and registers its routes.
func RegisterModule(
	roomRepo domain.RoomRepository,
	memberRepo domain.RoomMemberRepository,
	txManager shared_domain.TransactionManager,
	router *gin.RouterGroup,
	validate *validator.Validate,
	authMiddleware gin.HandlerFunc,
	listUserRoomsUseCase *application.ListUserRoomsUseCase, // Pass the use case in
) {
	// Initialize other use cases
	createRoomUseCase := application.NewCreateRoomUseCase(roomRepo, memberRepo, txManager)
	joinRoomUseCase := application.NewJoinRoomUseCase(roomRepo, memberRepo)

	// Initialize the controller, reusing the provided use case
	roomController := presentation.NewRoomController(
		createRoomUseCase,
		listUserRoomsUseCase,
		joinRoomUseCase,
		validate,
	)

	// Register routes under an authenticated group
	roomRoutes := router.Group("/rooms")
	roomRoutes.Use(authMiddleware)
	{
		roomRoutes.POST("/", roomController.CreateRoom)
		roomRoutes.GET("/", roomController.GetUserRooms)
		roomRoutes.POST("/:id/join", roomController.JoinRoom)
	}
}
