package repositories

import (
	"backend/domain"
	"context"
)

type MessageRepository interface {
	Create(ctx context.Context, message *domain.Message) (*domain.Message, error)
	FindByID(ctx context.Context, id string) (*domain.Message, error)
	FindByChannelID(ctx context.Context, channelID string) ([]*domain.Message, error)
}
