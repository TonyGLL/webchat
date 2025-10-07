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
	Text      string `json:"text"`
	AuthorID  string `json:"authorId"`
	ChannelID string `json:"channelId"`
}

func (uc *CreateMessageUseCase) Execute(input CreateMessageInputDTO) (*domain.Message, error) {
	message := &domain.Message{
		Text:      input.Text,
		AuthorID:  input.AuthorID,
		ChannelID: input.ChannelID,
		CreatedAt: time.Now(),
	}

	createdMessage, err := uc.MessageRepository.Create(message)
	if err != nil {
		return nil, err
	}

	return createdMessage, nil
}