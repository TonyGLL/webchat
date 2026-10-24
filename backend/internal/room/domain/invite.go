package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Invite struct {
	ID          uuid.UUID `json:"id"`
	RoomID      uuid.UUID `json:"room_id"`
	Code        string    `json:"code"`
	ExpiresAt   time.Time `json:"expires_at"`
	InvitedByID int       `json:"invited_by_id"`
	CreatedAt   time.Time `json:"created_at"`
	Email       string    `json:"email,omitempty"`
}

// ErrInviteNotFound is returned when the invite code does not exist or has expired.
var ErrInviteNotFound = errors.New("invite not found or expired")
