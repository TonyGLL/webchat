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

// MarkRead stamps last_read_at, but skips the write when it was stamped less than 2 s ago (BP-5):
// every viewer of a busy room sends one mark per incoming message and each write is a new row
// version. The member check is separate from the UPDATE so a throttled call still reports true.
func (r *PgRoomMemberRepository) MarkRead(ctx context.Context, roomID uuid.UUID, userID int) (bool, error) {
	var isMember bool
	err := r.db.QueryRowContext(ctx, `
		WITH touched AS (
			UPDATE room_members SET last_read_at = NOW()
			WHERE room_id = $1 AND user_id = $2 AND last_read_at < NOW() - interval '2 seconds'
			RETURNING 1
		)
		SELECT EXISTS (SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2)`, roomID, userID).Scan(&isMember)
	return isMember, err
}

func (r *PgRoomMemberRepository) RemoveMember(ctx context.Context, roomID uuid.UUID, userID int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM room_members WHERE room_id = $1 AND user_id = $2`, roomID, userID)
	return err
}

func (r *PgRoomMemberRepository) ListMembers(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*domain.MemberInfo, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT u.id, u.username, u.name, u.last_name, u.avatar_url, rm.role
		FROM room_members rm JOIN users u ON u.id = rm.user_id
		WHERE rm.room_id = $1 AND u.deleted = FALSE
		ORDER BY rm.joined_at, rm.user_id
		LIMIT $2 OFFSET $3`, roomID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	members := []*domain.MemberInfo{}
	for rows.Next() {
		m := &domain.MemberInfo{}
		if err := rows.Scan(&m.UserID, &m.Username, &m.Name, &m.LastName, &m.AvatarURL, &m.Role); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}
