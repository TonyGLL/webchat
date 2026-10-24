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

func (r *PgRepository) MessageRef(ctx context.Context, messageID uuid.UUID) (*domain.MessageRef, error) {
	ref := &domain.MessageRef{}
	var deletedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `SELECT room_id, author_id, deleted_at FROM messages WHERE id = $1`, messageID).Scan(&ref.RoomID, &ref.Author, &deletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrMessageNotFound
	}
	ref.Deleted = deletedAt.Valid
	return ref, err
}

func (r *PgRepository) SoftDeleteMessage(ctx context.Context, messageID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE messages SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, messageID)
	return err
}

const reportColumns = `rp.id, rp.reported_message_id, m.room_id, m.content, m.author_id, rp.reported_by_id, rp.reason, rp.status::text, rp.resolved_by_id, rp.created_at`

func scanReport(scan func(dest ...interface{}) error) (*domain.Report, error) {
	rep := &domain.Report{}
	err := scan(&rep.ID, &rep.MessageID, &rep.RoomID, &rep.MessageContent, &rep.MessageAuthor, &rep.ReportedBy, &rep.Reason, &rep.Status, &rep.ResolvedBy, &rep.CreatedAt)
	return rep, err
}

func (r *PgRepository) CreateReport(ctx context.Context, messageID uuid.UUID, reporterID int, reason string) (*domain.Report, error) {
	var id int
	err := r.db.QueryRowContext(ctx, `INSERT INTO reports (reported_message_id, reported_by_id, reason) VALUES ($1, $2, $3) RETURNING id`, messageID, reporterID, reason).Scan(&id)
	if db.IsUniqueViolation(err) {
		return nil, domain.ErrAlreadyReported
	}
	if err != nil {
		return nil, err
	}
	return r.FindReport(ctx, id)
}

func (r *PgRepository) FindReport(ctx context.Context, reportID int) (*domain.Report, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+reportColumns+` FROM reports rp JOIN messages m ON m.id = rp.reported_message_id WHERE rp.id = $1`, reportID)
	rep, err := scanReport(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrReportNotFound
	}
	return rep, err
}

func (r *PgRepository) ListReports(ctx context.Context, roomID uuid.UUID, status string, limit int) ([]*domain.Report, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+reportColumns+` FROM reports rp JOIN messages m ON m.id = rp.reported_message_id
		WHERE m.room_id = $1 AND rp.status = $2::report_status ORDER BY rp.created_at DESC LIMIT $3`, roomID, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	reports := []*domain.Report{}
	for rows.Next() {
		rep, err := scanReport(rows.Scan)
		if err != nil {
			return nil, err
		}
		reports = append(reports, rep)
	}
	return reports, rows.Err()
}

func (r *PgRepository) SetReportStatus(ctx context.Context, reportID int, status string, resolvedBy int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE reports SET status = $2::report_status, resolved_by_id = $3, updated_at = NOW() WHERE id = $1`, reportID, status, resolvedBy)
	return err
}

func (r *PgRepository) UpsertBan(ctx context.Context, ban domain.Ban) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO bans (user_id, room_id, banned_by_id, reason, expires_at) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, room_id) DO UPDATE
		SET banned_by_id = EXCLUDED.banned_by_id, reason = EXCLUDED.reason, expires_at = EXCLUDED.expires_at, created_at = NOW()`,
		ban.UserID, ban.RoomID, ban.BannedBy, ban.Reason, ban.ExpiresAt)
	return err
}

func (r *PgRepository) DeleteBan(ctx context.Context, roomID uuid.UUID, userID int) (bool, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM bans WHERE room_id = $1 AND user_id = $2`, roomID, userID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (r *PgRepository) ListBans(ctx context.Context, roomID uuid.UUID) ([]*domain.Ban, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT user_id, room_id, banned_by_id, COALESCE(reason, ''), expires_at, created_at
		FROM bans WHERE room_id = $1 AND (expires_at IS NULL OR expires_at > NOW()) ORDER BY created_at DESC`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	bans := []*domain.Ban{}
	for rows.Next() {
		b := &domain.Ban{}
		if err := rows.Scan(&b.UserID, &b.RoomID, &b.BannedBy, &b.Reason, &b.ExpiresAt, &b.CreatedAt); err != nil {
			return nil, err
		}
		bans = append(bans, b)
	}
	return bans, rows.Err()
}

func (r *PgRepository) IsBanned(ctx context.Context, roomID uuid.UUID, userID int) (bool, error) {
	var banned bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM bans WHERE room_id = $1 AND user_id = $2 AND (expires_at IS NULL OR expires_at > NOW()))`, roomID, userID).Scan(&banned)
	return banned, err
}
