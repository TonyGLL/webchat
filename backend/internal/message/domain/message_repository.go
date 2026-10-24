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
	// Search finds messages matching the text among the rooms the user belongs to (optionally
	// only in roomID), newest first. Deleted messages are never returned.
	Search(ctx context.Context, userID int, query string, roomID *uuid.UUID, limit int) ([]*Message, error)
	// MemberRole returns the user's role in the room and whether they are a member at all.
	MemberRole(ctx context.Context, roomID uuid.UUID, userID int) (role string, isMember bool, err error)
	SetPinned(ctx context.Context, id uuid.UUID, pinned bool) error
	// FindPinned returns the room's pinned, non-deleted messages, most recently pinned first.
	FindPinned(ctx context.Context, roomID uuid.UUID) ([]*Message, error)
	IsRoomMember(ctx context.Context, roomID uuid.UUID, userID int) (bool, error)
	// FindByRoomID returns up to limit messages, newest first. When before is set, only messages
	// older than that message are returned (cursor pagination, stable while new messages arrive).
	FindByRoomID(ctx context.Context, roomID uuid.UUID, limit int, before *uuid.UUID) ([]*Message, error)
}

type WebsocketBroadcaster interface {
	Broadcast(messageType string, payload interface{}, roomID string)
}
