package persistence

import (
	"backend/internal/message/domain"
	"backend/internal/shared/infra/db"
	"context"

	"github.com/google/uuid"
)

type PgMessageRepository struct {
	db db.DBTX
}

func NewPgMessageRepository(dbtx db.DBTX) domain.MessageRepository {
	return &PgMessageRepository{db: dbtx}
}

func (r *PgMessageRepository) Create(ctx context.Context, message *domain.Message) (*domain.Message, error) {
	query := "INSERT INTO messages (content, author_id, room_id) VALUES ($1, $2, $3) RETURNING id, created_at;"
	err := r.db.QueryRowContext(ctx, query, message.Content, message.AuthorID, message.RoomID).Scan(&message.ID, &message.CreatedAt)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (r *PgMessageRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Message, error) {
	query := `SELECT id, content, author_id, room_id, created_at, edited_at, deleted_at FROM messages WHERE id = `
	var msg domain.Message
	err := r.db.QueryRowContext(ctx, query, id).Scan(&msg.ID, &msg.Content, &msg.AuthorID, &msg.RoomID, &msg.CreatedAt, &msg.EditedAt, &msg.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *PgMessageRepository) FindByRoomID(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*domain.Message, error) {
	query := `SELECT id, content, author_id, room_id, created_at, edited_at, deleted_at FROM messages WHERE room_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, roomID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []*domain.Message
	for rows.Next() {
		var msg domain.Message
		err := rows.Scan(&msg.ID, &msg.Content, &msg.AuthorID, &msg.RoomID, &msg.CreatedAt, &msg.EditedAt, &msg.DeletedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}
