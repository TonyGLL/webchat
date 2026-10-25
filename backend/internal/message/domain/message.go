package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrMessageNotFound = errors.New("message not found")

// ErrSearchBusy: this instance already runs the maximum number of concurrent searches.
var ErrSearchBusy = errors.New("search is busy")

// ErrSearchTooBroad: the search did not finish within its time budget.
var ErrSearchTooBroad = errors.New("search took too long")

type Message struct {
	ID               uuid.UUID  `json:"id"`
	Content          string     `json:"content"`
	AuthorID         int        `json:"author_id"`
	RoomID           uuid.UUID  `json:"room_id"`
	ReplyToMessageID *uuid.UUID `json:"reply_to_message_id,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	EditedAt         *time.Time `json:"edited_at,omitempty"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
	PinnedAt         *time.Time `json:"pinned_at,omitempty"`
	Reactions        []Reaction `json:"reactions,omitempty"`
}
