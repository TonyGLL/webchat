
package domain

import (
	"context"

	"github.com/google/uuid"
)

type MessageRepository interface {
	Create(ctx context.Context, message *Message) (*Message, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Message, error)
	FindByRoomID(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*Message, error)
}

type WebsocketBroadcaster interface {
	Broadcast(messageType string, payload interface{}, roomID string)
}

