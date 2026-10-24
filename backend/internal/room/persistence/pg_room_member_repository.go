package persistence

import (
	"database/sql"
	"errors"

	"backend/internal/room/domain"
	"backend/internal/shared/infra/db"
	"context"

	"github.com/google/uuid"
)

type PgRoomMemberRepository struct {
	db db.DBTX
}

func NewPgRoomMemberRepository(dbtx db.DBTX) domain.RoomMemberRepository {
	return &PgRoomMemberRepository{db: dbtx}
}

func (r *PgRoomMemberRepository) AddMember(ctx context.Context, member *domain.RoomMember) error {
	query := `INSERT INTO room_members (room_id, user_id, role) VALUES ($1, $2, $3) ON CONFLICT (room_id, user_id) DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, member.RoomID, member.UserID, member.Role)
	return err
}

func (r *PgRoomMemberRepository) FindMember(ctx context.Context, roomID uuid.UUID, userID int) (*domain.RoomMember, error) {
	query := `SELECT room_id, user_id, role, joined_at FROM room_members WHERE room_id = $1 AND user_id = $2`
	member := &domain.RoomMember{}
	err := r.db.QueryRowContext(ctx, query, roomID, userID).Scan(&member.RoomID, &member.UserID, &member.Role, &member.JoinedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return member, nil
}

func (r *PgRoomMemberRepository) MarkRead(ctx context.Context, roomID uuid.UUID, userID int) (bool, error) {
	res, err := r.db.ExecContext(ctx, `UPDATE room_members SET last_read_at = NOW() WHERE room_id = $1 AND user_id = $2`, roomID, userID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (r *PgRoomMemberRepository) RemoveMember(ctx context.Context, roomID uuid.UUID, userID int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM room_members WHERE room_id = $1 AND user_id = $2`, roomID, userID)
	return err
}
