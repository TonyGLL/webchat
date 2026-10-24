package domain

import (
	"time"

	"github.com/google/uuid"
)

type RoomRole string

const (
	RoleOwner  RoomRole = "owner"
	RoleAdmin  RoomRole = "admin"
	RoleMember RoomRole = "member"
)

type Room struct {
	ID        uuid.UUID `json:"id"`
	OwnerID   int       `json:"owner_id"`
	Name      string    `json:"name"`
	Topic     *string   `json:"topic"`
	IsPrivate bool      `json:"is_private"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// UnreadCount is only filled when listing a user's rooms.
	UnreadCount int `json:"unread_count"`
	// MyRole is the requesting user's role; only filled when listing a user's rooms.
	MyRole RoomRole `json:"my_role,omitempty"`
}

type RoomMember struct {
	RoomID   uuid.UUID `json:"room_id"`
	UserID   int       `json:"user_id"`
	Role     RoomRole  `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}
