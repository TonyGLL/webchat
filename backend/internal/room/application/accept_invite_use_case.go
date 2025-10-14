package application

import (
	"backend/internal/room/domain"
	"context"
)

type AcceptInviteDTO struct {
	Code string `uri:"code" validate:"required"`
}

type AcceptInviteUseCase struct {
	inviteRepo     domain.InviteRepository
	roomMemberRepo domain.RoomMemberRepository
}

func NewAcceptInviteUseCase(inviteRepo domain.InviteRepository, roomMemberRepo domain.RoomMemberRepository) *AcceptInviteUseCase {
	return &AcceptInviteUseCase{inviteRepo: inviteRepo, roomMemberRepo: roomMemberRepo}
}

func (uc *AcceptInviteUseCase) Execute(ctx context.Context, dto AcceptInviteDTO, userID int) (*domain.Invite, error) {
	// Fetch the invite by code
	invite, err := uc.inviteRepo.GetByCode(ctx, dto.Code)
	if err != nil {
		return nil, err
	}

	// Add user to the room
	newMember := &domain.RoomMember{RoomID: invite.RoomID, UserID: userID, Role: domain.RoleMember}
	err = uc.roomMemberRepo.AddMember(ctx, newMember)
	if err != nil {
		return nil, err
	}

	return invite, nil
}
