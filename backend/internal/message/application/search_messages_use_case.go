package application

import (
	"context"
	"strings"

	"backend/internal/message/domain"
	shared_domain "backend/internal/shared/domain"

	"github.com/google/uuid"
)

const (
	minSearchLength    = 2
	maxSearchLength    = 100
	defaultSearchLimit = 20
	maxSearchLimit     = 50
)

// SearchMessagesUseCase finds messages by text, only in rooms the user belongs to.
type SearchMessagesUseCase struct {
	messageRepo domain.MessageRepository
}

func NewSearchMessagesUseCase(messageRepo domain.MessageRepository) *SearchMessagesUseCase {
	return &SearchMessagesUseCase{messageRepo: messageRepo}
}

// Execute searches every room of the user, or only roomIDStr when it is not empty.
func (uc *SearchMessagesUseCase) Execute(ctx context.Context, userID int, query, roomIDStr string, limit int) ([]*domain.Message, error) {
	query = strings.TrimSpace(query)
	if len([]rune(query)) < minSearchLength || len([]rune(query)) > maxSearchLength {
		return nil, shared_domain.ErrInvalidInput
	}
	if limit <= 0 {
		limit = defaultSearchLimit
	} else if limit > maxSearchLimit {
		limit = maxSearchLimit
	}

	var roomID *uuid.UUID
	if roomIDStr != "" {
		id, err := uuid.Parse(roomIDStr)
		if err != nil {
			return nil, shared_domain.ErrInvalidInput
		}
		if err := requireRoomMember(ctx, uc.messageRepo, id, userID); err != nil {
			return nil, err
		}
		roomID = &id
	}
	return uc.messageRepo.Search(ctx, userID, query, roomID, limit)
}
