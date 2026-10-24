package application

import (
	"context"

	"backend/internal/room/domain"
)

// DiscoverRoomsUseCase lists every room so people can find where to talk: public rooms can be
// joined straight away, private ones are listed but need an invitation to read or write.
type DiscoverRoomsUseCase struct {
	roomRepo domain.RoomRepository
}

func NewDiscoverRoomsUseCase(roomRepo domain.RoomRepository) *DiscoverRoomsUseCase {
	return &DiscoverRoomsUseCase{roomRepo: roomRepo}
}

func (uc *DiscoverRoomsUseCase) Execute(ctx context.Context, userID int) ([]*domain.RoomSummary, error) {
	return uc.roomRepo.ListDiscoverable(ctx, userID)
}
