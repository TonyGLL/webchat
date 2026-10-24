package domain

import (
	"context"

	"github.com/google/uuid"
)

type MessageRepository interface {
	Create(ctx context.Context, message *Message) (*Message, error)
	Update(ctx context.Context, content string, id uuid.UUID) (string, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Message, error)
	// SoftDelete marks the message as deleted and returns its room id.
	SoftDelete(ctx context.Context, id uuid.UUID) (string, error)
	IsRoomMember(ctx context.Context, roomID uuid.UUID, userID int) (bool, error)
	// FindByRoomID returns up to limit messages, newest first. When before is set, only messages
	// older than that message are returned (cursor pagination, stable while new messages arrive).
	FindByRoomID(ctx context.Context, roomID uuid.UUID, limit int, before *uuid.UUID) ([]*Message, error)
}

type WebsocketBroadcaster interface {
	Broadcast(messageType string, payload interface{}, roomID string)
}
