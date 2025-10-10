package application

import (
	"backend/internal/room/domain"
	"context"
)

type ListUserRoomsUseCase struct {
	roomRepo domain.RoomRepository
}

func NewListUserRoomsUseCase(roomRepo domain.RoomRepository) *ListUserRoomsUseCase {
	return &ListUserRoomsUseCase{roomRepo: roomRepo}
}

func (uc *ListUserRoomsUseCase) Execute(ctx context.Context, userID int) ([]*domain.Room, error) {
	return uc.roomRepo.FindRoomsByUserID(ctx, userID)
}
