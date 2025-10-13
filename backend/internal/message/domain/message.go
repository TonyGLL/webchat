package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrMessageNotFound = errors.New("message not found")

type Message struct {
	ID        uuid.UUID  `json:"id"`
	Content   string     `json:"content"`
	AuthorID  int        `json:"author_id"`
	RoomID    uuid.UUID  `json:"room_id"`
	CreatedAt time.Time  `json:"created_at"`
	EditedAt  *time.Time `json:"edited_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
