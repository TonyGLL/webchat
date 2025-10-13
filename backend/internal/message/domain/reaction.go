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

type ReactionRepository interface {
	Add(ctx context.Context, reaction *Reaction) error
	Remove(ctx context.Context, messageID uuid.UUID, userID int, emoji string) error
	FindByMessageID(ctx context.Context, messageID uuid.UUID) ([]*Reaction, error)
}
