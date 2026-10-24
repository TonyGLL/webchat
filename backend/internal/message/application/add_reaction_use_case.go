package application

import (
	"backend/internal/message/domain"
	shared_domain "backend/internal/shared/domain"
	"context"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	maxEmojiBytes          = 32 // the reactions.emoji column is VARCHAR(32)
	maxEmojiRunes          = 8  // enough for ZWJ sequences, rejects arbitrary text
	maxReactionsPerUser    = 10 // distinct emoji one user can put on a message
	maxReactionKindsPerMsg = 30 // distinct emoji a message can carry
)

type AddReactionDTO struct {
	Emoji string `json:"emoji" validate:"required,max=32"`
}

func validEmoji(emoji string) bool {
	return strings.TrimSpace(emoji) != "" && len(emoji) <= maxEmojiBytes && utf8.RuneCountInString(emoji) <= maxEmojiRunes
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
	msg, err := uc.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return err
	}
	if msg.DeletedAt != nil {
		return domain.ErrMessageNotFound
	}
	if err := requireRoomMember(ctx, uc.messageRepo, msg.RoomID, userID); err != nil {
		return err
	}
	if !validEmoji(dto.Emoji) {
		return shared_domain.ErrInvalidInput
	}
	counts, err := uc.reactionRepo.CountsFor(ctx, messageID, userID, dto.Emoji)
	if err != nil {
		return err
	}
	if !counts.AlreadyReacted {
		// a new reaction by this user: enforce the caps (re-adding an existing one is a no-op)
		if counts.UserReactions >= maxReactionsPerUser || (!counts.EmojiInUse && counts.Kinds >= maxReactionKindsPerMsg) {
			return shared_domain.ErrInvalidInput
		}
	}

	reaction := &domain.Reaction{MessageID: messageID, UserID: userID, Emoji: dto.Emoji}
	if err := uc.reactionRepo.Add(ctx, reaction); err != nil {
		return err
	}
	if uc.wsBroadcaster != nil {
		uc.wsBroadcaster.Broadcast("reaction_added", reaction, msg.RoomID.String())
	}
	return nil
}
