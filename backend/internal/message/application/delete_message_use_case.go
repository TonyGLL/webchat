package application

import (
	"context"
	"log/slog"

	"backend/internal/message/domain"
	shared_domain "backend/internal/shared/domain"

	"github.com/google/uuid"
)

// DeleteMessageUseCase soft-deletes a message. The author may delete their own; moderators and
// above may delete anyone's (and that is audited).
type DeleteMessageUseCase struct {
	messageRepo   domain.MessageRepository
	wsBroadcaster domain.WebsocketBroadcaster
	auditor       domain.Auditor
}

func NewDeleteMessageUseCase(messageRepo domain.MessageRepository, wsBroadcaster domain.WebsocketBroadcaster, auditor domain.Auditor) *DeleteMessageUseCase {
	return &DeleteMessageUseCase{messageRepo: messageRepo, wsBroadcaster: wsBroadcaster, auditor: auditor}
}

func (uc *DeleteMessageUseCase) Execute(ctx context.Context, messageID uuid.UUID, userID int) error {
	msg, err := uc.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return err
	}
	if msg.DeletedAt != nil {
		return domain.ErrMessageNotFound
	}
	byModerator := false
	if msg.AuthorID != userID {
		actor, err := uc.messageRepo.ActorIn(ctx, msg.RoomID, userID)
		if err != nil {
			return err
		}
		if !actor.CanDeleteAnyMessage() {
			return shared_domain.ErrForbidden
		}
		byModerator = true
	}

	roomID, err := uc.messageRepo.SoftDelete(ctx, messageID)
	if err != nil {
		return err
	}
	if byModerator && uc.auditor != nil {
		if err := uc.auditor.RecordMessageDeletedByModerator(ctx, userID, msg.RoomID, messageID, msg.AuthorID); err != nil {
			slog.Error("could not write audit log", "action", "message.deleted_by_moderator", "error", err)
		}
	}
	if uc.wsBroadcaster != nil {
		uc.wsBroadcaster.Broadcast("delete_message", map[string]interface{}{"message_id": messageID}, roomID)
	}
	return nil
}
