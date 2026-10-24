package application

import (
	"backend/internal/message/domain"
	shared_domain "backend/internal/shared/domain"
	"context"

	"github.com/google/uuid"
)

const (
	defaultMessageLimit = 50
	maxMessageLimit     = 100
)

type ListMessagesUseCase struct {
	messageRepo domain.MessageRepository
}

func NewListMessagesUseCase(messageRepo domain.MessageRepository) *ListMessagesUseCase {
	return &ListMessagesUseCase{messageRepo: messageRepo}
}

func (uc *ListMessagesUseCase) Execute(ctx context.Context, roomIDStr string, userID int, limit int, beforeStr string) ([]*domain.Message, error) {
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		return nil, shared_domain.ErrInvalidInput
	}

	if err := requireRoomMember(ctx, uc.messageRepo, roomID, userID); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = defaultMessageLimit
	} else if limit > maxMessageLimit {
		limit = maxMessageLimit
	}

	var before *uuid.UUID
	if beforeStr != "" {
		id, err := uuid.Parse(beforeStr)
		if err != nil {
			return nil, shared_domain.ErrInvalidInput
		}
		before = &id
	}

	return uc.messageRepo.FindByRoomID(ctx, roomID, limit, before)
}
