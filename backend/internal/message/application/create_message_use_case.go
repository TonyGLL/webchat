package application

import (
	"context"
	"strconv"

	"backend/internal/message/domain"
	shared_domain "backend/internal/shared/domain"
)

type CreateMessageUseCase struct {
	messageRepo domain.MessageRepository
}

func NewCreateMessageUseCase(messageRepo domain.MessageRepository) *CreateMessageUseCase {
	return &CreateMessageUseCase{messageRepo: messageRepo}
}

// Execute creates a new message. The author's ID is passed explicitly
// to ensure it's taken from a trusted source (like a JWT) and not from user input.
func (uc *CreateMessageUseCase) Execute(ctx context.Context, input CreateMessageDTO, authorID int) (*domain.Message, error) {
	channelID, err := strconv.Atoi(input.ChannelID)
	if err != nil {
		return nil, shared_domain.ErrInvalidInput // ChannelID should be a numeric string
	}

	message := &domain.Message{
		Text:      input.Text,
		AuthorID:  authorID,
		ChannelID: channelID,
		// CreatedAt is now handled by the database.
	}

	createdMessage, err := uc.messageRepo.Create(ctx, message)
	if err != nil {
		// In a real application, you might map specific database errors
		// to domain errors, e.g., a foreign key violation.
		return nil, err
	}

	return createdMessage, nil
}
