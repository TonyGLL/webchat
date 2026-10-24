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
			(SELECT COUNT(*) FROM messages m
			  WHERE m.room_id = r.id AND m.created_at > rm.last_read_at
			    AND m.author_id <> rm.user_id AND m.deleted_at IS NULL) AS unread_count,
			rm.role
		FROM rooms r JOIN room_members rm ON r.id = rm.room_id
		WHERE rm.user_id = $1 ORDER BY r.name`
	rows, err := r.db.QueryContext(ctx, query, userID)
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
