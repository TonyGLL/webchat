package persistence

import (
	"backend/application/repositories"
	"backend/domain"
	"context"
	"database/sql"
	"errors"
)

const (
	createQuery          = `INSERT INTO messages (text, author_id, channel_id) VALUES ($1, $2, $3) RETURNING id, created_at;`
	findByIDQuery        = `SELECT id, text, author_id, channel_id, created_at FROM messages WHERE id = $1;`
	findByChannelIDQuery = `SELECT id, text, author_id, channel_id, created_at FROM messages WHERE channel_id = $1 ORDER BY created_at ASC;`
)

// PgMessageRepository implements the repositories.MessageRepository interface for PostgreSQL.
type PgMessageRepository struct {
	db DBTX
}

// NewPgMessageRepository creates a new PgMessageRepository.
// It accepts a DBTX interface, which can be either a *sql.DB or *sql.Tx.
func NewPgMessageRepository(db DBTX) repositories.MessageRepository {
	return &PgMessageRepository{db: db}
}

func (r *PgMessageRepository) Create(ctx context.Context, message *domain.Message) (*domain.Message, error) {
	err := r.db.QueryRowContext(ctx, createQuery, message.Text, message.AuthorID, message.ChannelID).Scan(&message.ID, &message.CreatedAt)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (r *PgMessageRepository) FindByID(ctx context.Context, id int) (*domain.Message, error) {
	var message domain.Message
	err := r.db.QueryRowContext(ctx, findByIDQuery, id).Scan(&message.ID, &message.Text, &message.AuthorID, &message.ChannelID, &message.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &message, nil
}

func (r *PgMessageRepository) FindByChannelID(ctx context.Context, channelID int) ([]*domain.Message, error) {
	rows, err := r.db.QueryContext(ctx, findByChannelIDQuery, channelID)
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}