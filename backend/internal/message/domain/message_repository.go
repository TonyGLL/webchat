package domain

import (
	"backend/internal/shared/permissions"
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
	// ActorIn loads the user's room role (if a member) and global role, for permission checks.
	ActorIn(ctx context.Context, roomID uuid.UUID, userID int) (permissions.Actor, error)
	SetPinned(ctx context.Context, id uuid.UUID, pinned bool) error
	// FindPinned returns the room's pinned, non-deleted messages, most recently pinned first.
	FindPinned(ctx context.Context, roomID uuid.UUID) ([]*Message, error)
	IsRoomMember(ctx context.Context, roomID uuid.UUID, userID int) (bool, error)
	// FindByRoomID returns up to limit messages, newest first. When before is set, only messages
	// older than that message are returned (cursor pagination, stable while new messages arrive).
	FindByRoomID(ctx context.Context, roomID uuid.UUID, limit int, before *uuid.UUID) ([]*Message, error)
}

// Auditor records moderation actions taken on messages.
type Auditor interface {
	RecordMessageDeletedByModerator(ctx context.Context, actorID int, roomID, messageID uuid.UUID, authorID int) error
}

type WebsocketBroadcaster interface {
	Broadcast(messageType string, payload interface{}, roomID string)
}
