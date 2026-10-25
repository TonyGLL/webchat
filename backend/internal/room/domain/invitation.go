package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrAlreadyInvited = errors.New("this user already has a pending invitation to the room")

// Invitation is a request addressed to one user to join a room.
type Invitation struct {
	ID          uuid.UUID `json:"id"`
	RoomID      uuid.UUID `json:"room_id"`
	RoomName    string    `json:"room_name"`
	InviterID   int       `json:"inviter_id"`
	InviterName string    `json:"inviter_name"`
	InviteeID   int       `json:"invitee_id"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	// InviteeName is filled only by ListByRoom (the admin view).
	InviteeName string `json:"invitee_name,omitempty"`
}

type InvitationRepository interface {
	// Create stores a pending invitation; ErrAlreadyInvited if one is already open, ErrUserNotFound if the invitee does not exist.
	Create(ctx context.Context, roomID uuid.UUID, inviterID, inviteeID int) (*Invitation, error)
	// ListByRoom returns the open (unexpired) invitations of a room, newest first.
	ListByRoom(ctx context.Context, roomID uuid.UUID) ([]*Invitation, error)
	// Cancel closes a pending invitation of the room; false if there was none.
	Cancel(ctx context.Context, roomID, id uuid.UUID) (bool, error)
	// ListPending returns the open (unexpired) invitations addressed to userID, newest first.
	ListPending(ctx context.Context, userID int) ([]*Invitation, error)
	// FindPending returns a pending, unexpired invitation, or nil.
	FindPending(ctx context.Context, id uuid.UUID) (*Invitation, error)
	// Resolve closes a pending invitation as accepted or declined; false if it was no longer pending.
	Resolve(ctx context.Context, id uuid.UUID, accepted bool) (bool, error)
}
