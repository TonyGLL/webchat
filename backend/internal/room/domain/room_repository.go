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
	// FindRoomsByUserID lists the user's rooms with their unread count (capped at MaxUnreadCount).
	FindRoomsByUserID(ctx context.Context, userID int) ([]*Room, error)
	// FindRoomIDsByUserID is the light version of FindRoomsByUserID: only the ids, no unread counts.
	FindRoomIDsByUserID(ctx context.Context, userID int) ([]uuid.UUID, error)
	// ListDiscoverable returns one page of non-direct rooms ordered by (name, id), flagging the ones
	// userID already belongs to. after, when set, is the id of the last room of the previous page.
	ListDiscoverable(ctx context.Context, userID, limit int, after *uuid.UUID) ([]*RoomSummary, error)
	// FindActiveUser returns the user, or nil when it does not exist or is deactivated.
	FindActiveUser(ctx context.Context, id int) (*UserRef, error)
}

type RoomMemberRepository interface {
	AddMember(ctx context.Context, member *RoomMember) error
	FindMember(ctx context.Context, roomID uuid.UUID, userID int) (*RoomMember, error)
	// MarkRead sets the member's last-read time to now. It reports false if they are not a member.
	// ListMembers returns a page of the members of a room, oldest first (Online is left false; see
	// PresenceChecker).
	ListMembers(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*MemberInfo, error)
	RemoveMember(ctx context.Context, roomID uuid.UUID, userID int) error
	MarkRead(ctx context.Context, roomID uuid.UUID, userID int) (bool, error)
}
