package persistence

import (
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
	query := `SELECT id, owner_id, name, topic, is_private, created_at, updated_at FROM rooms WHERE id = $1`
	room := &domain.Room{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&room.ID, &room.OwnerID, &room.Name, &room.Topic, &room.IsPrivate, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return room, nil
}

func (r *PgRoomRepository) FindRoomsByUserID(ctx context.Context, userID int) ([]*domain.Room, error) {
	query := `SELECT r.id, r.owner_id, r.name, r.topic, r.is_private, r.created_at, r.updated_at FROM rooms r JOIN room_members rm ON r.id = rm.room_id WHERE rm.user_id = $1 ORDER BY r.name`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rooms []*domain.Room
	for rows.Next() {
		room := &domain.Room{}
		err := rows.Scan(&room.ID, &room.OwnerID, &room.Name, &room.Topic, &room.IsPrivate, &room.CreatedAt, &room.UpdatedAt)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}
