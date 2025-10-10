package websocket

import "encoding/json"

// Message represents a message to be broadcasted to a room.
// This is an internal representation used by the Hub.
type Message struct {
	// Type indicates the nature of the message (e.g., "new_message", "user_joined").
	Type string
	// Payload is the actual data being sent.
	Payload interface{}
	// RoomID is the identifier of the room to which the message should be sent.
	RoomID string
}

// WebsocketMessage is the structure of the message sent to the client.
// This is the public-facing structure.
type WebsocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
