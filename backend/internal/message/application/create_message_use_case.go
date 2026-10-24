package application

import (
	"backend/internal/message/domain"
	shared_domain "backend/internal/shared/domain"
	"context"

	"github.com/google/uuid"
)

type CreateMessageDTO struct {
	Content          string  `json:"content" validate:"required,min=1,max=4000"`
	RoomID           string  `json:"room_id" validate:"required,uuid"`
	ReplyToMessageID *string `json:"reply_to_message_id,omitempty" validate:"omitempty,uuid"`
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
	if err := requireRoomMember(ctx, uc.messageRepo, roomID, authorID); err != nil {
		return nil, err
	}

	message := &domain.Message{Content: input.Content, AuthorID: authorID, RoomID: roomID}
	if input.ReplyToMessageID != nil {
		replyID, err := uuid.Parse(*input.ReplyToMessageID)
		if err != nil {
			return nil, shared_domain.ErrInvalidInput
		}
		parent, err := uc.messageRepo.FindByID(ctx, replyID)
		if err != nil || parent.DeletedAt != nil || parent.RoomID != roomID {
			// only replies to existing messages of the same room are allowed
			return nil, shared_domain.ErrInvalidInput
		}
		message.ReplyToMessageID = &replyID
	}
	createdMessage, err := uc.messageRepo.Create(ctx, message)
	if err != nil {
		return nil, err
	}
	if uc.wsBroadcaster != nil {
		uc.wsBroadcaster.Broadcast("new_message", createdMessage, input.RoomID)
	}
	return createdMessage, nil
}
