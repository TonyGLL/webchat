package domain

import (
	"context"
)

type MessageRepository interface {
	Create(ctx context.Context, message *Message) (*Message, error)
	FindByID(ctx context.Context, id int) (*Message, error)
	FindByChannelID(ctx context.Context, channelID int) ([]*Message, error)
}