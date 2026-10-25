package persistence

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"backend/internal/room/domain"
	"backend/internal/shared/infra/db"

	"github.com/google/uuid"
)

type PgInvitationRepository struct {
	db db.DBTX
}

func NewPgInvitationRepository(dbtx db.DBTX) domain.InvitationRepository {
	return &PgInvitationRepository{db: dbtx}
}

// maxPendingInvitations bounds the pending list so a flood of invitations cannot make it unbounded.
const maxPendingInvitations = 100

const invitationSelect = `
	SELECT i.id, i.room_id, r.name, i.inviter_id, u.name || COALESCE(' ' || u.last_name, ''), i.invitee_id, i.created_at
	FROM room_invitations i
	JOIN rooms r ON r.id = i.room_id
	JOIN users u ON u.id = i.inviter_id`

func scanInvitation(row interface{ Scan(...any) error }) (*domain.Invitation, error) {
	inv := &domain.Invitation{}
	err := row.Scan(&inv.ID, &inv.RoomID, &inv.RoomName, &inv.InviterID, &inv.InviterName, &inv.InviteeID, &inv.CreatedAt)
	return inv, err
}

func (r *PgInvitationRepository) Create(ctx context.Context, roomID uuid.UUID, inviterID, inviteeID int) (*domain.Invitation, error) {
	var id uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO room_invitations (room_id, inviter_id, invitee_id) VALUES ($1, $2, $3) RETURNING id`,
		roomID, inviterID, inviteeID).Scan(&id)
	switch {
	case db.IsUniqueViolation(err):
		return nil, domain.ErrAlreadyInvited
	case db.IsForeignKeyViolation(err):
		return nil, domain.ErrUserNotFound
	case err != nil:
		return nil, err
	}
	return scanInvitation(r.db.QueryRowContext(ctx, invitationSelect+` WHERE i.id = $1`, id))
}

func (r *PgInvitationRepository) ListPending(ctx context.Context, userID int) ([]*domain.Invitation, error) {
	rows, err := r.db.QueryContext(ctx, invitationSelect+` WHERE i.invitee_id = $1 AND i.status = 'pending' ORDER BY i.created_at DESC LIMIT `+strconv.Itoa(maxPendingInvitations), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []*domain.Invitation{}
	for rows.Next() {
		inv, err := scanInvitation(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, inv)
	}
	return list, rows.Err()
}

func (r *PgInvitationRepository) FindPending(ctx context.Context, id uuid.UUID) (*domain.Invitation, error) {
	inv, err := scanInvitation(r.db.QueryRowContext(ctx, invitationSelect+` WHERE i.id = $1 AND i.status = 'pending'`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return inv, err
}

func (r *PgInvitationRepository) Resolve(ctx context.Context, id uuid.UUID, accepted bool) (bool, error) {
	status := "declined"
	if accepted {
		status = "accepted"
	}
	res, err := r.db.ExecContext(ctx,
		`UPDATE room_invitations SET status = $2, responded_at = NOW() WHERE id = $1 AND status = 'pending'`, id, status)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}
