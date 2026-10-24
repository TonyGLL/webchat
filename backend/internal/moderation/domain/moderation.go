// Package domain holds the moderation types: who may do what is decided by shared/permissions;
// this module records and applies the decisions.
package domain

import (
	"context"
	"errors"
	"time"

	"backend/internal/shared/permissions"

	"github.com/google/uuid"
)

var (
	// ErrNotMember is returned when the target of an action is not in the room.
	ErrNotMember = errors.New("the user is not a member of the room")
	// ErrCannotTargetOwner is returned for any attempt to change or remove the room owner.
	ErrCannotTargetOwner = errors.New("the room owner cannot be changed or removed")
)

// Audit actions.
const (
	ActionRoleChanged    = "member.role_changed"
	ActionMemberKicked   = "member.kicked"
	ActionMemberBanned   = "member.banned"
	ActionMemberUnbanned = "member.unbanned"
	ActionMessageDeleted = "message.deleted_by_moderator"
	ActionReportResolved = "report.resolved"
	ActionReportIgnored  = "report.ignored"
)

// AuditEntry is one row of the append-only audit log.
type AuditEntry struct {
	ID              int64                  `json:"id"`
	ActorID         *int                   `json:"actor_id"`
	Action          string                 `json:"action"`
	RoomID          *uuid.UUID             `json:"room_id,omitempty"`
	TargetUserID    *int                   `json:"target_user_id,omitempty"`
	TargetMessageID *uuid.UUID             `json:"target_message_id,omitempty"`
	Details         map[string]interface{} `json:"details"`
	CreatedAt       time.Time              `json:"created_at"`
}

// Repository is the persistence the moderation use cases need.
type Repository interface {
	// ActorIn loads a user's room role (if a member) and global role.
	ActorIn(ctx context.Context, roomID uuid.UUID, userID int) (permissions.Actor, error)
	SetMemberRole(ctx context.Context, roomID uuid.UUID, userID int, role permissions.RoomRole) error
	RemoveMember(ctx context.Context, roomID uuid.UUID, userID int) error
	AddAudit(ctx context.Context, entry AuditEntry) error
	ListAudit(ctx context.Context, roomID uuid.UUID, limit int) ([]*AuditEntry, error)
	// RecordMessageDeletedByModerator lets the message module audit moderator deletions.
	RecordMessageDeletedByModerator(ctx context.Context, actorID int, roomID, messageID uuid.UUID, authorID int) error
}

// Realtime lets moderation actions reach connected clients.
type Realtime interface {
	Broadcast(messageType string, payload interface{}, roomID string)
	UnsubscribeUser(userID int, roomID string)
}
