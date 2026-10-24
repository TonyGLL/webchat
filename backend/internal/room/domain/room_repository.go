package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrUserNotFound = errors.New("user not found")

type RoomRepository interface {
	Create(ctx context.Context, room *Room) (*Room, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Room, error)
	// FindOrCreateDirect returns the one-to-one room between two users, creating it if needed.
	// ErrUserNotFound is returned when the other user does not exist.
	FindOrCreateDirect(ctx context.Context, userID, otherID int) (*Room, error)
	FindRoomsByUserID(ctx context.Context, userID int) ([]*Room, error)
}

type RoomMemberRepository interface {
	AddMember(ctx context.Context, member *RoomMember) error
	FindMember(ctx context.Context, roomID uuid.UUID, userID int) (*RoomMember, error)
	// MarkRead sets the member's last-read time to now. It reports false if they are not a member.
	// ListMembers returns the members of a room (Online is left false; see PresenceChecker).
	ListMembers(ctx context.Context, roomID uuid.UUID) ([]*MemberInfo, error)
	RemoveMember(ctx context.Context, roomID uuid.UUID, userID int) error
	MarkRead(ctx context.Context, roomID uuid.UUID, userID int) (bool, error)
}
