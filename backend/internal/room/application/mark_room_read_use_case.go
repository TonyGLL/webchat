package application

import (
	"context"

	"backend/internal/room/domain"
	shared_domain "backend/internal/shared/domain"

	"github.com/google/uuid"
)

// MarkRoomReadUseCase records that a member has seen everything in a room so far.
type MarkRoomReadUseCase struct {
	memberRepo domain.RoomMemberRepository
}

func NewMarkRoomReadUseCase(memberRepo domain.RoomMemberRepository) *MarkRoomReadUseCase {
	return &MarkRoomReadUseCase{memberRepo: memberRepo}
}

func (uc *MarkRoomReadUseCase) Execute(ctx context.Context, roomID uuid.UUID, userID int) error {
	marked, err := uc.memberRepo.MarkRead(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if !marked {
		return shared_domain.ErrForbidden // not a member of that room
	}
	return nil
}
