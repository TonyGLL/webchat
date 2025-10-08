package usecases

import (
	"backend/application"
	"backend/application/dtos"
	"backend/domain"
	"context"
	"strconv"
)

type CreateMessageUseCase struct {
	store application.Store
}

func NewCreateMessageUseCase(store application.Store) *CreateMessageUseCase {
	return &CreateMessageUseCase{store: store}
}

// Execute creates a new message. The author's ID is passed explicitly
// to ensure it's taken from a trusted source (like a JWT) and not from user input.
func (uc *CreateMessageUseCase) Execute(ctx context.Context, input dtos.CreateMessageDTO, authorID int) (*domain.Message, error) {
	channelID, err := strconv.Atoi(input.ChannelID)
	if err != nil {
		return nil, domain.ErrInvalidInput // ChannelID should be a numeric string
	}

	message := &domain.Message{
		Text:      input.Text,
		AuthorID:  authorID,
		ChannelID: channelID,
		// CreatedAt is now handled by the database.
	}

	createdMessage, err := uc.store.MessageRepository().Create(ctx, message)
	if err != nil {
		// In a real application, you might map specific database errors
		// to domain errors, e.g., a foreign key violation.
		return nil, err
	}

	return createdMessage, nil
}