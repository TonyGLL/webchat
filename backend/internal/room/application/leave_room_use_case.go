package application

import (
	"context"
	"errors"

	"backend/internal/room/domain"
	shared_domain "backend/internal/shared/domain"

	"github.com/google/uuid"
)

// ErrOwnerCannotLeave is returned when the room owner tries to leave: the room would be left without an owner.
var ErrOwnerCannotLeave = errors.New("the room owner cannot leave the room")

type LeaveRoomUseCase struct {
	memberRepo domain.RoomMemberRepository
	subscriber domain.RealtimeSubscriber
}

func NewLeaveRoomUseCase(memberRepo domain.RoomMemberRepository, subscriber domain.RealtimeSubscriber) *LeaveRoomUseCase {
	return &LeaveRoomUseCase{memberRepo: memberRepo, subscriber: subscriber}
}

func (uc *LeaveRoomUseCase) Execute(ctx context.Context, roomID uuid.UUID, userID int) error {
	member, err := uc.memberRepo.FindMember(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return shared_domain.ErrForbidden
	}
	if member.Role == domain.RoleOwner {
		return ErrOwnerCannotLeave
	}
	if err := uc.memberRepo.RemoveMember(ctx, roomID, userID); err != nil {
		return err
	}
	if uc.subscriber != nil {
		uc.subscriber.UnsubscribeUser(userID, roomID.String())
	}
	return nil
}
