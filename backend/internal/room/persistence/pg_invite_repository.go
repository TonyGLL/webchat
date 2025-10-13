//go:build !test

package persistence

import (
	"backend/internal/room/domain"
	"backend/internal/shared/infra/db"
	"context"
)

// NOTE: This repository requires a new `invites` table.
// You should create a migration with the following SQL:
// CREATE TABLE invites (
//     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//     room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
//     code TEXT NOT NULL UNIQUE,
//     expires_at TIMESTAMPTZ NOT NULL,
//     created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
//     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
// );

type PgInviteRepository struct {
	db db.DBTX
}

func NewPgInviteRepository(dbtx db.DBTX) domain.InviteRepository {
	return &PgInviteRepository{db: dbtx}
}

func (r *PgInviteRepository) Create(ctx context.Context, invite *domain.Invite) (*domain.Invite, error) {
	query := `INSERT INTO invites (room_id, code, expires_at, invited_by_id, email) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	err := r.db.QueryRowContext(ctx, query, invite.RoomID, invite.Code, invite.ExpiresAt, invite.InvitedByID, invite.Email).Scan(&invite.ID, &invite.CreatedAt)
	if err != nil {
		return nil, err
	}
	return invite, nil
}
