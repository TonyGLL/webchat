package domain

// RealtimeSubscriber subscribes a user's open websocket connections to a room, so
// rooms created or joined after connecting receive events without reconnecting.
type RealtimeSubscriber interface {
	SubscribeUser(userID int, roomID string)
}
