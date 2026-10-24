package persistence

import (
	"database/sql"
	"errors"

	"backend/internal/room/domain"
	"backend/internal/shared/infra/db"
	"context"

	"github.com/google/uuid"
)

type PgRoomRepository struct {
	db db.DBTX
}

func NewPgRoomRepository(dbtx db.DBTX) domain.RoomRepository {
	return &PgRoomRepository{db: dbtx}
}

func (r *PgRoomRepository) Create(ctx context.Context, room *domain.Room) (*domain.Room, error) {
	query := `INSERT INTO rooms (owner_id, name, topic, is_private) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	err := r.db.QueryRowContext(ctx, query, room.OwnerID, room.Name, room.Topic, room.IsPrivate).Scan(&room.ID, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return room, nil
}

func (r *PgRoomRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
	query := `SELECT id, owner_id, name, topic, is_private, is_direct, created_at, updated_at FROM rooms WHERE id = $1`
	room := &domain.Room{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&room.ID, &room.OwnerID, &room.Name, &room.Topic, &room.IsPrivate, &room.IsDirect, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return room, nil
}

func (r *PgRoomRepository) FindRoomsByUserID(ctx context.Context, userID int) ([]*domain.Room, error) {
	query := `
		SELECT r.id, r.owner_id,
			CASE WHEN r.is_direct THEN COALESCE(
				(SELECT u.name || COALESCE(' ' || u.last_name, '') FROM room_members o JOIN users u ON u.id = o.user_id
				  WHERE o.room_id = r.id AND o.user_id <> rm.user_id LIMIT 1), 'Deleted user')
			ELSE r.name END AS name,
			r.topic, r.is_private, r.is_direct, r.created_at, r.updated_at,
			(SELECT COUNT(*) FROM (
				SELECT 1 FROM messages m
				 WHERE m.room_id = r.id AND m.created_at > rm.last_read_at
				   AND m.author_id IS DISTINCT FROM rm.user_id AND m.deleted_at IS NULL
				 LIMIT $2) capped) AS unread_count,
			rm.role
		FROM rooms r JOIN room_members rm ON r.id = rm.room_id
		WHERE rm.user_id = $1 ORDER BY r.name`
	rows, err := r.db.QueryContext(ctx, query, userID, domain.MaxUnreadCount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rooms := []*domain.Room{}
	for rows.Next() {
		room := &domain.Room{}
		err := rows.Scan(&room.ID, &room.OwnerID, &room.Name, &room.Topic, &room.IsPrivate, &room.IsDirect, &room.CreatedAt, &room.UpdatedAt, &room.UnreadCount, &room.MyRole)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (r *PgRoomRepository) FindRoomIDsByUserID(ctx context.Context, userID int) ([]uuid.UUID, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT room_id FROM room_members WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []uuid.UUID{}
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// ListDiscoverable pages with keyset on (name, id). The member count stays a correlated subquery:
// with the LIMIT it only runs for the rows of the page, so a join would not be cheaper.
func (r *PgRoomRepository) ListDiscoverable(ctx context.Context, userID, limit int, after *uuid.UUID) ([]*domain.RoomSummary, error) {
	query := `
		SELECT r.id, r.name, r.topic, r.is_private,
			(SELECT COUNT(*) FROM room_members m WHERE m.room_id = r.id) AS member_count,
			EXISTS (SELECT 1 FROM room_members m WHERE m.room_id = r.id AND m.user_id = $1) AS is_member
		FROM rooms r
		WHERE NOT r.is_direct
		  AND ($2::uuid IS NULL OR (r.name, r.id) > (SELECT a.name, a.id FROM rooms a WHERE a.id = $2::uuid))
		ORDER BY r.name, r.id
		LIMIT $3`
	var cursor any
	if after != nil {
		cursor = after.String()
	}
	rows, err := r.db.QueryContext(ctx, query, userID, cursor, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rooms := []*domain.RoomSummary{}
	for rows.Next() {
		room := &domain.RoomSummary{}
		if err := rows.Scan(&room.ID, &room.Name, &room.Topic, &room.IsPrivate, &room.MemberCount, &room.IsMember); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, rows.Err()
}

func (r *PgRoomRepository) FindActiveUser(ctx context.Context, id int) (*domain.UserRef, error) {
	u := &domain.UserRef{}
	err := r.db.QueryRowContext(ctx, `SELECT id, email FROM users WHERE id = $1 AND deleted = FALSE`, id).Scan(&u.ID, &u.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

const directRoomQuery = `SELECT r.id, r.owner_id, r.name, r.topic, r.is_private, r.is_direct, r.created_at, r.updated_at
	FROM direct_rooms d JOIN rooms r ON r.id = d.room_id WHERE d.user_a = $1 AND d.user_b = $2`

// createDirectRoomQuery creates the room, both memberships and the pair record in ONE statement,
// so it either happens completely or not at all (and a concurrent duplicate fails on direct_rooms).
const createDirectRoomQuery = `
	WITH r AS (
		INSERT INTO rooms (owner_id, name, is_private, is_direct) VALUES ($1::int, 'direct', TRUE, TRUE)
		RETURNING id, owner_id, name, topic, is_private, is_direct, created_at, updated_at
	), m AS (
		INSERT INTO room_members (room_id, user_id, role) SELECT id, $2::int, 'member'::room_role FROM r
		UNION ALL SELECT id, $3::int, 'member'::room_role FROM r
	), d AS (
		INSERT INTO direct_rooms (user_a, user_b, room_id) SELECT $2::int, $3::int, id FROM r
	)
	SELECT id, owner_id, name, topic, is_private, is_direct, created_at, updated_at FROM r`

func scanRoom(row *sql.Row) (*domain.Room, error) {
	room := &domain.Room{}
	err := row.Scan(&room.ID, &room.OwnerID, &room.Name, &room.Topic, &room.IsPrivate, &room.IsDirect, &room.CreatedAt, &room.UpdatedAt)
	return room, err
}

func (r *PgRoomRepository) FindOrCreateDirect(ctx context.Context, userID, otherID int) (*domain.Room, error) {
	a, b := userID, otherID
	if a > b {
		a, b = b, a
	}

	room, err := scanRoom(r.db.QueryRowContext(ctx, directRoomQuery, a, b))
	if err == nil {
		return room, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	room, err = scanRoom(r.db.QueryRowContext(ctx, createDirectRoomQuery, userID, a, b))
	switch {
	case err == nil:
		return room, nil
	case db.IsForeignKeyViolation(err):
		return nil, domain.ErrUserNotFound
	case db.IsUniqueViolation(err):
		// someone created it at the same moment: use theirs
		return scanRoom(r.db.QueryRowContext(ctx, directRoomQuery, a, b))
	default:
		return nil, err
	}
}
