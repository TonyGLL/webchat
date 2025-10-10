package websocket

import (
	"context"
)

// RoomInfo represents basic information about a room.
type RoomInfo struct {
	ID string
}

// RoomLister defines the interface for listing a user's rooms.
// This allows the websocket handler to be decoupled from the room module's application layer.
type RoomLister interface {
	GetUserRoomIDs(ctx context.Context, userID int) ([]string, error)
}
