package domain

import (
	"context"

	"github.com/google/uuid"
)

type RoomRepository interface {
	Create(ctx context.Context, room *Room) (*Room, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Room, error)
	FindRoomsByUserID(ctx context.Context, userID int) ([]*Room, error)
}

type RoomMemberRepository interface {
	AddMember(ctx context.Context, member *RoomMember) error
	FindMember(ctx context.Context, roomID uuid.UUID, userID int) (*RoomMember, error)
	// MarkRead sets the member's last-read time to now. It reports false if they are not a member.
	RemoveMember(ctx context.Context, roomID uuid.UUID, userID int) error
	MarkRead(ctx context.Context, roomID uuid.UUID, userID int) (bool, error)
}
