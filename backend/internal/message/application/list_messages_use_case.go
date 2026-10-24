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

func (uc *ListMessagesUseCase) Execute(ctx context.Context, roomIDStr string, userID int, page, pageSize int) ([]*domain.Message, error) {
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		return nil, shared_domain.ErrInvalidInput
	}

	if err := requireRoomMember(ctx, uc.messageRepo, roomID, userID); err != nil {
		return nil, err
	}

	if pageSize <= 0 {
		pageSize = defaultMessageLimit
	} else if pageSize > maxMessageLimit {
		pageSize = maxMessageLimit
	}

	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * pageSize

	return uc.messageRepo.FindByRoomID(ctx, roomID, pageSize, offset)
}
