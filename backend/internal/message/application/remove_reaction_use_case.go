package application

import (
	"backend/internal/message/domain"
	"context"

	"github.com/google/uuid"
)

type RemoveReactionDTO struct {
	Emoji string `json:"emoji" validate:"required"`
}

type RemoveReactionUseCase struct {
	reactionRepo  domain.ReactionRepository
	messageRepo   domain.MessageRepository
	wsBroadcaster domain.WebsocketBroadcaster
}

func NewRemoveReactionUseCase(reactionRepo domain.ReactionRepository, messageRepo domain.MessageRepository, wsBroadcaster domain.WebsocketBroadcaster) *RemoveReactionUseCase {
	return &RemoveReactionUseCase{reactionRepo: reactionRepo, messageRepo: messageRepo, wsBroadcaster: wsBroadcaster}
}

func (uc *RemoveReactionUseCase) Execute(ctx context.Context, messageID uuid.UUID, userID int, dto RemoveReactionDTO) error {
	if err := uc.reactionRepo.Remove(ctx, messageID, userID, dto.Emoji); err != nil {
		return err
	}
	msg, err := uc.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return err
	}
	if msg == nil {
		return domain.ErrMessageNotFound
	}
	if uc.wsBroadcaster != nil {
		payload := map[string]interface{}{"message_id": messageID, "user_id": userID, "emoji": dto.Emoji}
		uc.wsBroadcaster.Broadcast("reaction_removed", payload, msg.RoomID.String())
	}
	return nil
}
