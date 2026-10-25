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
	// ErrReportNotFound / ErrMessageNotFound are returned for ids that do not exist (or are gone).
	ErrReportNotFound  = errors.New("report not found")
	ErrMessageNotFound = errors.New("message not found")
	// ErrAlreadyReported is returned when the same user reports the same message twice.
	ErrAlreadyReported = errors.New("you already reported this message")
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

// Report statuses.
const (
	ReportPending  = "pending"
	ReportResolved = "resolved"
	ReportIgnored  = "ignored"
)

// Report is a member's complaint about a message.
type Report struct {
	ID             int       `json:"id"`
	MessageID      uuid.UUID `json:"message_id"`
	RoomID         uuid.UUID `json:"room_id"`
	MessageContent string    `json:"message_content"`
	MessageAuthor  *int      `json:"message_author_id"`
	ReportedBy     int       `json:"reported_by_id"`
	Reason         string    `json:"reason"`
	Status         string    `json:"status"`
	ResolvedBy     *int      `json:"resolved_by_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// MessageRef is the little we need to know about a message to act on it.
type MessageRef struct {
	RoomID  uuid.UUID
	Author  *int
	Deleted bool
}

// Ban keeps a user out of a room, forever (ExpiresAt nil) or until a date.
type Ban struct {
	UserID    int        `json:"user_id"`
	RoomID    uuid.UUID  `json:"room_id"`
	BannedBy  *int       `json:"banned_by_id"`
	Reason    string     `json:"reason"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	// BannerRank is the issuer's rank when the ban was set; lifting or overwriting the ban needs a
	// higher one (or being the issuer). Zero for legacy rows whose issuer had no power left.
	BannerRank permissions.Rank `json:"-"`
}

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
	MessageRef(ctx context.Context, messageID uuid.UUID) (*MessageRef, error)
	SoftDeleteMessage(ctx context.Context, messageID uuid.UUID) error
	CreateReport(ctx context.Context, messageID uuid.UUID, reporterID int, reason string) (*Report, error)
	FindReport(ctx context.Context, reportID int) (*Report, error)
	ListReports(ctx context.Context, roomID uuid.UUID, status string, limit int) ([]*Report, error)
	SetReportStatus(ctx context.Context, reportID int, status string, resolvedBy int) error
	UpsertBan(ctx context.Context, ban Ban) error
	// FindBan returns the active (unexpired) ban of the user in the room, or nil.
	FindBan(ctx context.Context, roomID uuid.UUID, userID int) (*Ban, error)
	DeleteBan(ctx context.Context, roomID uuid.UUID, userID int) (bool, error)
	ListBans(ctx context.Context, roomID uuid.UUID) ([]*Ban, error)
	// IsBanned reports whether the user has an active (unexpired) ban in the room.
	IsBanned(ctx context.Context, roomID uuid.UUID, userID int) (bool, error)
	// RecordMessageDeletedByModerator lets the message module audit moderator deletions.
	RecordMessageDeletedByModerator(ctx context.Context, actorID int, roomID, messageID uuid.UUID, authorID int) error
}

// Realtime lets moderation actions reach connected clients.
type Realtime interface {
	Broadcast(messageType string, payload interface{}, roomID string)
	UnsubscribeUser(userID int, roomID string)
}
