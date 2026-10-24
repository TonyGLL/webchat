package application

import (
	"context"
	"strings"
	"unicode/utf8"

	"backend/internal/message/domain"
	shared_domain "backend/internal/shared/domain"

	"github.com/google/uuid"
)

// requireRoomMember returns ErrForbidden unless the user belongs to the room.
func requireRoomMember(ctx context.Context, repo domain.MessageRepository, roomID uuid.UUID, userID int) error {
	isMember, err := repo.IsRoomMember(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return shared_domain.ErrForbidden
	}
	return nil
}

// maxContentLength is the longest message body, in characters.
const maxContentLength = 4000

// validateContent rejects blank or over-long message content (shared by create and edit).
func validateContent(content string) error {
	if strings.TrimSpace(content) == "" || utf8.RuneCountInString(content) > maxContentLength {
		return shared_domain.ErrInvalidInput
	}
	return nil
}
