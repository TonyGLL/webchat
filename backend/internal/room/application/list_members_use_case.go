package application

import (
	"context"

	"backend/internal/room/domain"
	shared_domain "backend/internal/shared/domain"

	"github.com/google/uuid"
)

// ListMembersUseCase lists a room's members with their online status. Only members may see it.
type ListMembersUseCase struct {
	memberRepo domain.RoomMemberRepository
	presence   domain.PresenceChecker
}

func NewListMembersUseCase(memberRepo domain.RoomMemberRepository, presence domain.PresenceChecker) *ListMembersUseCase {
	return &ListMembersUseCase{memberRepo: memberRepo, presence: presence}
}

func (uc *ListMembersUseCase) Execute(ctx context.Context, roomID uuid.UUID, userID int) ([]*domain.MemberInfo, error) {
	self, err := uc.memberRepo.FindMember(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if self == nil {
		return nil, shared_domain.ErrForbidden
	}

	members, err := uc.memberRepo.ListMembers(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if uc.presence == nil {
		return members, nil
	}

	ids := make([]int, len(members))
	for i, m := range members {
		ids[i] = m.UserID
	}
	online, err := uc.presence.OnlineUsers(ctx, ids)
	if err != nil {
		return members, nil // presence is a nicety: show the members even if it is unavailable
	}
	for _, m := range members {
		m.Online = online[m.UserID]
	}
	return members, nil
}
