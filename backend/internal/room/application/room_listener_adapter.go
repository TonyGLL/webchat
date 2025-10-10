package application

import (
	"context"
)

type RoomListerAdapter struct {
	useCase *ListUserRoomsUseCase
}

func NewRoomListerAdapter(useCase *ListUserRoomsUseCase) *RoomListerAdapter {
	return &RoomListerAdapter{useCase: useCase}
}

func (a *RoomListerAdapter) GetUserRoomIDs(ctx context.Context, userID int) ([]string, error) {
	rooms, err := a.useCase.Execute(ctx, userID)
	if err != nil {
		return nil, err
	}
	roomIDs := make([]string, len(rooms))
	for i, room := range rooms {
		roomIDs[i] = room.ID.String()
	}
	return roomIDs, nil
}
