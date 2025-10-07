package usecases

import (
	"backend/application/repositories"
	"backend/domain"
	"time"
)

type CreateMessageUseCase struct {
	MessageRepository repositories.MessageRepository
}

func NewCreateMessageUseCase(repo repositories.MessageRepository) *CreateMessageUseCase {
	return &CreateMessageUseCase{MessageRepository: repo}
}

type CreateMessageInputDTO struct {
	Text      string `json:"text" validate:"required"`
	AuthorID  string `json:"authorId" validate:"required"`
	ChannelID string `json:"channelId" validate:"required"`
}

func (uc *CreateMessageUseCase) Execute(input CreateMessageInputDTO) (*domain.Message, error) {
	if input.Text == "" {
		return nil, domain.ErrInvalidInput
	}

	message := &domain.Message{
		Text:      input.Text,
		AuthorID:  input.AuthorID,
		ChannelID: input.ChannelID,
		CreatedAt: time.Now(),
	}

	createdMessage, err := uc.MessageRepository.Create(message)
	if err != nil {
		// In a real application, you might want to check for specific database errors
		// and map them to domain errors, e.g., a unique constraint violation
		// could be mapped to domain.ErrConflict.
		return nil, domain.ErrInternal
	}

	return createdMessage, nil
}