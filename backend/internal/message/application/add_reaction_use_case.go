package application

import (
	"backend/internal/message/domain"
	"context"

	"github.com/google/uuid"
)

type AddReactionDTO struct {
	Emoji string `json:"emoji" validate:"required"`
}

type AddReactionUseCase struct {
	reactionRepo  domain.ReactionRepository
	messageRepo   domain.MessageRepository
	wsBroadcaster domain.WebsocketBroadcaster
}

func NewAddReactionUseCase(reactionRepo domain.ReactionRepository, messageRepo domain.MessageRepository, wsBroadcaster domain.WebsocketBroadcaster) *AddReactionUseCase {
	return &AddReactionUseCase{reactionRepo: reactionRepo, messageRepo: messageRepo, wsBroadcaster: wsBroadcaster}
}

func (uc *AddReactionUseCase) Execute(ctx context.Context, messageID uuid.UUID, userID int, dto AddReactionDTO) error {
	reaction := &domain.Reaction{MessageID: messageID, UserID: userID, Emoji: dto.Emoji}
	if err := uc.reactionRepo.Add(ctx, reaction); err != nil {
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
		uc.wsBroadcaster.Broadcast("reaction_added", reaction, msg.RoomID.String())
	}
	return nil
}
