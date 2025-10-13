package persistence

import (
	"backend/internal/message/domain"
	"backend/internal/shared/infra/db"
	"context"

	"github.com/google/uuid"
)

type PgReactionRepository struct {
	db db.DBTX
}

func NewPgReactionRepository(dbtx db.DBTX) domain.ReactionRepository {
	return &PgReactionRepository{db: dbtx}
}

func (r *PgReactionRepository) Add(ctx context.Context, reaction *domain.Reaction) error {
	query := `INSERT INTO reactions (message_id, user_id, emoji) VALUES ($1, $2, $3) ON CONFLICT (message_id, user_id, emoji) DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, reaction.MessageID, reaction.UserID, reaction.Emoji)
	return err
}

func (r *PgReactionRepository) Remove(ctx context.Context, messageID uuid.UUID, userID int, emoji string) error {
	query := `DELETE FROM reactions WHERE message_id = $1 AND user_id = $2 AND emoji = $3`
	_, err := r.db.ExecContext(ctx, query, messageID, userID, emoji)
	return err
}

func (r *PgReactionRepository) FindByMessageID(ctx context.Context, messageID uuid.UUID) ([]*domain.Reaction, error) {
	query := `SELECT message_id, user_id, emoji, created_at FROM reactions WHERE message_id = $1 ORDER BY created_at ASC`
	rows, err := r.db.QueryContext(ctx, query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var reactions []*domain.Reaction
	for rows.Next() {
		var reaction domain.Reaction
		if err := rows.Scan(&reaction.MessageID, &reaction.UserID, &reaction.Emoji, &reaction.CreatedAt); err != nil {
			return nil, err
		}
		reactions = append(reactions, &reaction)
	}
	return reactions, nil
}
