package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Reaction struct {
	MessageID uuid.UUID `json:"message_id"`
	UserID    int       `json:"user_id"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}

// ReactionCounts summarises a message's reactions from one user's point of view.
type ReactionCounts struct {
	UserReactions  int  // distinct emoji this user has put on the message
	Kinds          int  // distinct emoji on the message overall
	EmojiInUse     bool // somebody already used the emoji being added
	AlreadyReacted bool // this user already used the emoji being added
}

type ReactionRepository interface {
	CountsFor(ctx context.Context, messageID uuid.UUID, userID int, emoji string) (ReactionCounts, error)
	Add(ctx context.Context, reaction *Reaction) error
	Remove(ctx context.Context, messageID uuid.UUID, userID int, emoji string) error
	FindByMessageID(ctx context.Context, messageID uuid.UUID) ([]*Reaction, error)
}
