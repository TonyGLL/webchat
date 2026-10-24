package application

import (
	"backend/internal/message/domain"
	shared_domain "backend/internal/shared/domain"
	"context"

	"github.com/google/uuid"
)

type UpdateMessageUseCase struct {
	messageRepo   domain.MessageRepository
	wsBroadcaster domain.WebsocketBroadcaster
}

type UpdateMessageDTO struct {
	Content string `json:"content" validate:"required"`
}

func NewUpdateMessageUseCase(messageRepo domain.MessageRepository, wsBroadcaster domain.WebsocketBroadcaster) *UpdateMessageUseCase {
	return &UpdateMessageUseCase{
		messageRepo:   messageRepo,
		wsBroadcaster: wsBroadcaster,
	}
}

func (uc *UpdateMessageUseCase) Execute(ctx context.Context, messageID uuid.UUID, userID int, dto UpdateMessageDTO) error {
	msg, err := uc.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return err
	}
	if msg.AuthorID != userID {
		return shared_domain.ErrForbidden
	}

	roomID, err := uc.messageRepo.Update(ctx, dto.Content, messageID)
	if err != nil {
		return err
	}

	if uc.wsBroadcaster != nil {
		uc.wsBroadcaster.Broadcast("update_message", dto, roomID)
	}
	return nil
}
