package application

import (
	"context"

	"backend/internal/message/domain"
	shared_domain "backend/internal/shared/domain"

	"github.com/google/uuid"
)

// DeleteMessageUseCase soft-deletes a message; only its author may do it.
type DeleteMessageUseCase struct {
	messageRepo   domain.MessageRepository
	wsBroadcaster domain.WebsocketBroadcaster
}

func NewDeleteMessageUseCase(messageRepo domain.MessageRepository, wsBroadcaster domain.WebsocketBroadcaster) *DeleteMessageUseCase {
	return &DeleteMessageUseCase{messageRepo: messageRepo, wsBroadcaster: wsBroadcaster}
}

func (uc *DeleteMessageUseCase) Execute(ctx context.Context, messageID uuid.UUID, userID int) error {
	msg, err := uc.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return err
	}
	if msg.DeletedAt != nil {
		return domain.ErrMessageNotFound
	}
	if msg.AuthorID != userID {
		return shared_domain.ErrForbidden
	}

	roomID, err := uc.messageRepo.SoftDelete(ctx, messageID)
	if err != nil {
		return err
	}
	if uc.wsBroadcaster != nil {
		uc.wsBroadcaster.Broadcast("delete_message", map[string]interface{}{"message_id": messageID}, roomID)
	}
	return nil
}
