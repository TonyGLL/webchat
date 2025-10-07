package persistence

import (
	"context"
	"errors"

	"backend/application/repositories"
	"backend/domain"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	createQuery         = `INSERT INTO messages (text, author_id, channel_id) VALUES ($1, $2, $3) RETURNING id, created_at;`
	findByIDQuery       = `SELECT id, text, author_id, channel_id, created_at FROM messages WHERE id = $1;`
	findByChannelIDQuery = `SELECT id, text, author_id, channel_id, created_at FROM messages WHERE channel_id = $1 ORDER BY created_at ASC;`
)

type PgMessageRepository struct {
	pool *pgxpool.Pool
}

func NewPgMessageRepository(pool *pgxpool.Pool) repositories.MessageRepository {
	return &PgMessageRepository{pool: pool}
}

func (r *PgMessageRepository) Create(message *domain.Message) (*domain.Message, error) {
	err := r.pool.QueryRow(context.Background(), createQuery, message.Text, message.AuthorID, message.ChannelID).Scan(&message.ID, &message.CreatedAt)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (r *PgMessageRepository) FindByID(id string) (*domain.Message, error) {
	var message domain.Message
	err := r.pool.QueryRow(context.Background(), findByIDQuery, id).Scan(&message.ID, &message.Text, &message.AuthorID, &message.ChannelID, &message.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &message, nil
}

func (r *PgMessageRepository) FindByChannelID(channelID string) ([]*domain.Message, error) {
	rows, err := r.pool.Query(context.Background(), findByChannelIDQuery, channelID)
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