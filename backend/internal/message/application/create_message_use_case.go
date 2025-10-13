package application

import (
	"backend/internal/message/domain"
	shared_domain "backend/internal/shared/domain"
	"context"

	"github.com/google/uuid"
)

type CreateMessageDTO struct {
	Content string `json:"content" validate:"required,min=1,max=4000"`
	RoomID  string `json:"room_id" validate:"required,uuid"`
}

type CreateMessageUseCase struct {
	messageRepo   domain.MessageRepository
	wsBroadcaster domain.WebsocketBroadcaster
}

func NewCreateMessageUseCase(messageRepo domain.MessageRepository, wsBroadcaster domain.WebsocketBroadcaster) *CreateMessageUseCase {
	return &CreateMessageUseCase{messageRepo: messageRepo, wsBroadcaster: wsBroadcaster}
}

func (uc *CreateMessageUseCase) Execute(ctx context.Context, input CreateMessageDTO, authorID int) (*domain.Message, error) {
	roomID, err := uuid.Parse(input.RoomID)
	if err != nil {
		return nil, shared_domain.ErrInvalidInput
	}
	message := &domain.Message{Content: input.Content, AuthorID: authorID, RoomID: roomID}
	createdMessage, err := uc.messageRepo.Create(ctx, message)
	if err != nil {
		return nil, err
	}
	if uc.wsBroadcaster != nil {
		uc.wsBroadcaster.Broadcast("new_message", createdMessage, input.RoomID)
	}
	return createdMessage, nil
}
