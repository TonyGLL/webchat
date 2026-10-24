package persistence

import (
	"backend/internal/message/domain"
	"backend/internal/shared/infra/db"
	"context"
	"database/sql"
	"encoding/json"
	"errors"

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

func (r *PgMessageRepository) SoftDelete(ctx context.Context, id uuid.UUID) (string, error) {
	var roomID string
	err := r.db.QueryRowContext(ctx, `UPDATE messages SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL RETURNING room_id`, id).Scan(&roomID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", domain.ErrMessageNotFound
	}
	return roomID, err
}

func (r *PgMessageRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Message, error) {
	query := `SELECT id, content, author_id, room_id, created_at, edited_at, deleted_at FROM messages WHERE id = $1;`
	var msg domain.Message
	err := r.db.QueryRowContext(ctx, query, id).Scan(&msg.ID, &msg.Content, &msg.AuthorID, &msg.RoomID, &msg.CreatedAt, &msg.EditedAt, &msg.DeletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrMessageNotFound
		}
		return nil, err
	}
	return &msg, nil
}

func (r *PgMessageRepository) IsRoomMember(ctx context.Context, roomID uuid.UUID, userID int) (bool, error) {
	var isMember bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2);`, roomID, userID).Scan(&isMember)
	return isMember, err
}

func (r *PgMessageRepository) FindByRoomID(ctx context.Context, roomID uuid.UUID, limit int, before *uuid.UUID) ([]*domain.Message, error) {
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
			m.pinned_at,
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
		  AND m.deleted_at IS NULL
		  AND ($2::uuid IS NULL OR (m.created_at, m.id) < (SELECT created_at, id FROM messages WHERE id = $2))
		GROUP BY m.id
		ORDER BY m.created_at DESC, m.id DESC
		LIMIT $3
	`
	rows, err := r.db.QueryContext(ctx, query, roomID, before, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	messages := []*domain.Message{}
	for rows.Next() {
		var msg domain.Message
		var reactionsJSON []byte
		err := rows.Scan(&msg.ID, &msg.Content, &msg.AuthorID, &msg.RoomID, &msg.ReplyToMessageID, &msg.CreatedAt, &msg.EditedAt, &msg.DeletedAt, &msg.PinnedAt, &reactionsJSON)
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

func (r *PgMessageRepository) MemberRole(ctx context.Context, roomID uuid.UUID, userID int) (string, bool, error) {
	var role string
	err := r.db.QueryRowContext(ctx, `SELECT role::text FROM room_members WHERE room_id = $1 AND user_id = $2`, roomID, userID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	return role, err == nil, err
}

func (r *PgMessageRepository) SetPinned(ctx context.Context, id uuid.UUID, pinned bool) error {
	_, err := r.db.ExecContext(ctx, `UPDATE messages SET pinned_at = CASE WHEN $2 THEN NOW() ELSE NULL END WHERE id = $1`, id, pinned)
	return err
}

func (r *PgMessageRepository) FindPinned(ctx context.Context, roomID uuid.UUID) ([]*domain.Message, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, content, author_id, room_id, reply_to_message_id, created_at, edited_at, deleted_at, pinned_at
		FROM messages
		WHERE room_id = $1 AND pinned_at IS NOT NULL AND deleted_at IS NULL
		ORDER BY pinned_at DESC`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	messages := []*domain.Message{}
	for rows.Next() {
		var msg domain.Message
		if err := rows.Scan(&msg.ID, &msg.Content, &msg.AuthorID, &msg.RoomID, &msg.ReplyToMessageID, &msg.CreatedAt, &msg.EditedAt, &msg.DeletedAt, &msg.PinnedAt); err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}
	return messages, rows.Err()
}

func (r *PgMessageRepository) Search(ctx context.Context, userID int, query string, roomID *uuid.UUID, limit int) ([]*domain.Message, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT m.id, m.content, m.author_id, m.room_id, m.reply_to_message_id, m.created_at, m.edited_at, m.deleted_at, m.pinned_at
		FROM messages m
		JOIN room_members rm ON rm.room_id = m.room_id AND rm.user_id = $1
		WHERE m.deleted_at IS NULL
		  AND to_tsvector('simple', search_unaccent(m.content)) @@ websearch_to_tsquery('simple', search_unaccent($2))
		  AND ($3::uuid IS NULL OR m.room_id = $3)
		ORDER BY m.created_at DESC
		LIMIT $4`, userID, query, roomID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	messages := []*domain.Message{}
	for rows.Next() {
		var msg domain.Message
		if err := rows.Scan(&msg.ID, &msg.Content, &msg.AuthorID, &msg.RoomID, &msg.ReplyToMessageID, &msg.CreatedAt, &msg.EditedAt, &msg.DeletedAt, &msg.PinnedAt); err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}
	return messages, rows.Err()
}
