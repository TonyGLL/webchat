package domain

import (
	"context"

	"github.com/google/uuid"
)

// RealtimeSubscriber subscribes a user's open websocket connections to a room, so
// rooms created or joined after connecting receive events without reconnecting.
type RealtimeSubscriber interface {
	SubscribeUser(userID int, roomID string)
	// UnsubscribeUser stops delivering a room's events to the user's open connections.
	UnsubscribeUser(userID int, roomID string)
}

// PresenceChecker tells which users currently have an open connection.
type PresenceChecker interface {
	OnlineUsers(ctx context.Context, userIDs []int) (map[int]bool, error)
}

// BanChecker tells whether a user is currently banned from a room.
type BanChecker interface {
	IsBanned(ctx context.Context, roomID uuid.UUID, userID int) (bool, error)
}
