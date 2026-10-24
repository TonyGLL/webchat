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
	subscriber     domain.RealtimeSubscriber
	bans           domain.BanChecker
}

func NewAcceptInviteUseCase(inviteRepo domain.InviteRepository, roomMemberRepo domain.RoomMemberRepository, subscriber domain.RealtimeSubscriber, bans domain.BanChecker) *AcceptInviteUseCase {
	return &AcceptInviteUseCase{inviteRepo: inviteRepo, roomMemberRepo: roomMemberRepo, subscriber: subscriber, bans: bans}
}

func (uc *AcceptInviteUseCase) Execute(ctx context.Context, dto AcceptInviteDTO, userID int) (*domain.Invite, error) {
	// Fetch the invite by code
	invite, err := uc.inviteRepo.GetByCode(ctx, dto.Code)
	if err != nil {
		return nil, err
	}

	if uc.bans != nil {
		banned, err := uc.bans.IsBanned(ctx, invite.RoomID, userID)
		if err != nil {
			return nil, err
		}
		if banned {
			return nil, ErrBannedFromRoom
		}
	}

	member, err := uc.roomMemberRepo.FindMember(ctx, invite.RoomID, userID)
	if err != nil {
		return nil, err
	}
	if member != nil {
		return nil, ErrAlreadyMember
	}

	// Add user to the room
	newMember := &domain.RoomMember{RoomID: invite.RoomID, UserID: userID, Role: domain.RoleMember}
	err = uc.roomMemberRepo.AddMember(ctx, newMember)
	if err != nil {
		return nil, err
	}

	if uc.subscriber != nil {
		uc.subscriber.SubscribeUser(userID, invite.RoomID.String())
	}

	return invite, nil
}
