package usecases

import (
	"backend/application/dtos"
	"backend/application/repositories"
	"backend/domain"
	"context"
	"strconv"
	"time"
)

type CreateMessageUseCase struct {
	MessageRepository repositories.MessageRepository
}

func NewCreateMessageUseCase(repo repositories.MessageRepository) *CreateMessageUseCase {
	return &CreateMessageUseCase{MessageRepository: repo}
}

// Execute creates a new message. The author's ID is passed explicitly
// to ensure it's taken from a trusted source (like a JWT) and not from user input.
func (uc *CreateMessageUseCase) Execute(ctx context.Context, input dtos.CreateMessageDTO, authorID int) (*domain.Message, error) {
	message := &domain.Message{
		Text:      input.Text,
		AuthorID:  strconv.Itoa(authorID), // Assuming AuthorID in domain.Message is a string
		ChannelID: input.ChannelID,
		CreatedAt: time.Now(), // The database can also handle this with a default value
	}

	createdMessage, err := uc.MessageRepository.Create(ctx, message)
	if err != nil {
		// In a real application, you might want to check for specific database errors
		// and map them to domain errors, e.g., a unique constraint violation
		// could be mapped to domain.ErrConflict.
		return nil, err
	}

	return createdMessage, nil
}
