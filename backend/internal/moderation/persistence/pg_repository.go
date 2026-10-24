package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"backend/internal/moderation/domain"
	"backend/internal/shared/infra/db"
	"backend/internal/shared/permissions"

	"github.com/google/uuid"
)

type PgRepository struct {
	db db.DBTX
}

func NewPgRepository(dbtx db.DBTX) domain.Repository {
	return &PgRepository{db: dbtx}
}

func (r *PgRepository) ActorIn(ctx context.Context, roomID uuid.UUID, userID int) (permissions.Actor, error) {
	var global string
	err := r.db.QueryRowContext(ctx, `SELECT role FROM users WHERE id = $1 AND deleted = FALSE`, userID).Scan(&global)
	if errors.Is(err, sql.ErrNoRows) {
		return permissions.Actor{}, nil // unknown user: no power at all
	}
	if err != nil {
		return permissions.Actor{}, err
	}
	actor := permissions.Actor{Global: permissions.GlobalRole(global)}

	var room string
	err = r.db.QueryRowContext(ctx, `SELECT role::text FROM room_members WHERE room_id = $1 AND user_id = $2`, roomID, userID).Scan(&room)
	switch {
	case err == nil:
		actor.Room, actor.IsMember = permissions.RoomRole(room), true
	case !errors.Is(err, sql.ErrNoRows):
		return permissions.Actor{}, err
	}
	return actor, nil
}

func (r *PgRepository) SetMemberRole(ctx context.Context, roomID uuid.UUID, userID int, role permissions.RoomRole) error {
	_, err := r.db.ExecContext(ctx, `UPDATE room_members SET role = $3::room_role WHERE room_id = $1 AND user_id = $2`, roomID, userID, string(role))
	return err
}

func (r *PgRepository) RemoveMember(ctx context.Context, roomID uuid.UUID, userID int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM room_members WHERE room_id = $1 AND user_id = $2`, roomID, userID)
	return err
}

func (r *PgRepository) AddAudit(ctx context.Context, e domain.AuditEntry) error {
	details, err := json.Marshal(e.Details)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO audit_log (actor_id, action, room_id, target_user_id, target_message_id, details) VALUES ($1, $2, $3, $4, $5, $6)`,
		e.ActorID, e.Action, e.RoomID, e.TargetUserID, e.TargetMessageID, details)
	return err
}

func (r *PgRepository) ListAudit(ctx context.Context, roomID uuid.UUID, limit int) ([]*domain.AuditEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, actor_id, action, room_id, target_user_id, target_message_id, details, created_at
		FROM audit_log WHERE room_id = $1 ORDER BY created_at DESC, id DESC LIMIT $2`, roomID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	entries := []*domain.AuditEntry{}
	for rows.Next() {
		e := &domain.AuditEntry{}
		var details []byte
		if err := rows.Scan(&e.ID, &e.ActorID, &e.Action, &e.RoomID, &e.TargetUserID, &e.TargetMessageID, &details, &e.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(details, &e.Details); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// RecordMessageDeletedByModerator lets the message module audit moderator deletions
// (it satisfies message/domain.Auditor without importing it).
func (r *PgRepository) RecordMessageDeletedByModerator(ctx context.Context, actorID int, roomID, messageID uuid.UUID, authorID int) error {
	return r.AddAudit(ctx, domain.AuditEntry{
		ActorID: &actorID, Action: domain.ActionMessageDeleted, RoomID: &roomID, TargetMessageID: &messageID,
		TargetUserID: &authorID, Details: map[string]interface{}{},
	})
}
