package application

import (
	"context"

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
