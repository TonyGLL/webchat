package persistence

import (
	"context"
	"database/sql"
	"errors"

	"backend/application/repositories"
	"backend/domain"
)

const (
	createQuery          = `INSERT INTO messages (text, author_id, channel_id) VALUES ($1, $2, $3) RETURNING id, created_at;`
	findByIDQuery        = `SELECT id, text, author_id, channel_id, created_at FROM messages WHERE id = $1;`
	findByChannelIDQuery = `SELECT id, text, author_id, channel_id, created_at FROM messages WHERE channel_id = $1 ORDER BY created_at ASC;`
)

type PgMessageRepository struct {
	queries *Queries
}

func NewPgMessageRepository(db *sql.DB) repositories.MessageRepository {
	return &PgMessageRepository{queries: New(db)}
}

func (r *PgMessageRepository) Create(ctx context.Context, message *domain.Message) (*domain.Message, error) {
	err := r.queries.db.QueryRowContext(ctx, createQuery, message.Text, message.AuthorID, message.ChannelID).Scan(&message.ID, &message.CreatedAt)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (r *PgMessageRepository) FindByID(ctx context.Context, id string) (*domain.Message, error) {
	var message domain.Message
	err := r.queries.db.QueryRowContext(ctx, findByIDQuery, id).Scan(&message.ID, &message.Text, &message.AuthorID, &message.ChannelID, &message.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &message, nil
}

func (r *PgMessageRepository) FindByChannelID(ctx context.Context, channelID string) ([]*domain.Message, error) {
	rows, err := r.queries.db.QueryContext(ctx, findByChannelIDQuery, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		var message domain.Message
		err := rows.Scan(&message.ID, &message.Text, &message.AuthorID, &message.ChannelID, &message.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &message)
	}

	return messages, nil
}
