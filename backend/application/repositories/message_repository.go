package repositories

import "backend/domain"

type MessageRepository interface {
	Create(message *domain.Message) (*domain.Message, error)
	FindByID(id string) (*domain.Message, error)
	FindByChannelID(channelID string) ([]*domain.Message, error)
}
