package persistence

import (
	"backend/internal/message/domain"
	"backend/internal/shared/infra/db"
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

type PgMessageRepository struct {
	db db.DBTX
}

func NewPgMessageRepository(dbtx db.DBTX) domain.MessageRepository {
	return &PgMessageRepository{db: dbtx}
}

func (r *PgMessageRepository) Create(ctx context.Context, message *domain.Message) (*domain.Message, error) {
	query := "INSERT INTO messages (content, author_id, room_id, reply_to_message_id) VALUES ($1, $2, $3, $4) RETURNING id, created_at;"
	err := r.db.QueryRowContext(ctx, query, message.Content, message.AuthorID, message.RoomID, message.ReplyToMessageID).Scan(&message.ID, &message.CreatedAt)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (r *PgMessageRepository) Update(ctx context.Context, content string, id uuid.UUID) (string, error) {
	query := "UPDATE messages SET content = $1, edited_at = NOW() WHERE id = $2 returning room_id"
	var roomID string
	err := r.db.QueryRowContext(ctx, query, content, id).Scan(&roomID)
	if err != nil {
		return "", err
	}
	return roomID, nil
}

func (r *PgMessageRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Message, error) {
	query := `SELECT id, content, author_id, room_id, created_at, edited_at, deleted_at FROM messages WHERE id = $1;`
	var msg domain.Message
	err := r.db.QueryRowContext(ctx, query, id).Scan(&msg.ID, &msg.Content, &msg.AuthorID, &msg.RoomID, &msg.CreatedAt, &msg.EditedAt, &msg.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *PgMessageRepository) FindByRoomID(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*domain.Message, error) {
	query := `
		SELECT 
			m.id, 
			m.content, 
			m.author_id, 
			m.room_id, 
			m.reply_to_message_id, 
			m.created_at, 
			m.edited_at, 
			m.deleted_at,
			COALESCE(json_agg(
				json_build_object(
					'message_id', r.message_id,
					'user_id', r.user_id,
					'emoji', r.emoji,
					'created_at', r.created_at
				)
			) FILTER (WHERE r.message_id IS NOT NULL), '[]') AS reactions
		FROM messages m
		LEFT JOIN reactions r ON m.id = r.message_id
		WHERE m.room_id = $1
		GROUP BY m.id
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, roomID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []*domain.Message
	for rows.Next() {
		var msg domain.Message
		var reactionsJSON []byte
		err := rows.Scan(&msg.ID, &msg.Content, &msg.AuthorID, &msg.RoomID, &msg.ReplyToMessageID, &msg.CreatedAt, &msg.EditedAt, &msg.DeletedAt, &reactionsJSON)
		if err != nil {
			return nil, err
		}

		// Unmarshal al tipo correcto
		var reactions []domain.Reaction
		if len(reactionsJSON) > 0 {
			if err := json.Unmarshal(reactionsJSON, &reactions); err != nil {
				return nil, err
			}
		}
		msg.Reactions = reactions

		messages = append(messages, &msg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}
