package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"backend/internal/message/domain"
	shared_domain "backend/internal/shared/domain"

	"github.com/google/uuid"
)

const (
	minSearchLength    = 2
	maxSearchLength    = 100
	defaultSearchLimit = 20
	maxSearchLimit     = 50

	// searchTimeout bounds one search (BP-2): a common word can scan every message of every
	// room of the user, and lib/pq cancels the statement server-side when the context ends.
	searchTimeout = 2 * time.Second
	// maxConcurrentSearches caps searches running at once per instance so a few of them cannot
	// hold the whole DB pool.
	maxConcurrentSearches = 4
)

// SearchMessagesUseCase finds messages by text, only in rooms the user belongs to.
type SearchMessagesUseCase struct {
	messageRepo domain.MessageRepository
	slots       chan struct{}
	timeout     time.Duration
}

func NewSearchMessagesUseCase(messageRepo domain.MessageRepository) *SearchMessagesUseCase {
	return &SearchMessagesUseCase{
		messageRepo: messageRepo,
		slots:       make(chan struct{}, maxConcurrentSearches),
		timeout:     searchTimeout,
	}
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

	// Do not queue: when the instance is saturated, fail fast so the client can retry later.
	select {
	case uc.slots <- struct{}{}:
		defer func() { <-uc.slots }()
	default:
		return nil, domain.ErrSearchBusy
	}

	searchCtx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()
	messages, err := uc.messageRepo.Search(searchCtx, userID, query, roomID, limit)
	if err != nil && errors.Is(searchCtx.Err(), context.DeadlineExceeded) && ctx.Err() == nil {
		return nil, domain.ErrSearchTooBroad
	}
	return messages, err
}
