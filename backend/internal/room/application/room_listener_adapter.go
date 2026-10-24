package application

import (
	"context"

	"backend/internal/room/domain"
)

// RoomListerAdapter lets the websocket layer look up a user's rooms without importing this module.
// It only needs the ids, so it skips the unread counts of the full room list.
type RoomListerAdapter struct {
	roomRepo domain.RoomRepository
}

func NewRoomListerAdapter(roomRepo domain.RoomRepository) *RoomListerAdapter {
	return &RoomListerAdapter{roomRepo: roomRepo}
}

func (a *RoomListerAdapter) GetUserRoomIDs(ctx context.Context, userID int) ([]string, error) {
	ids, err := a.roomRepo.FindRoomIDsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	roomIDs := make([]string, len(ids))
	for i, id := range ids {
		roomIDs[i] = id.String()
	}
	return roomIDs, nil
}
