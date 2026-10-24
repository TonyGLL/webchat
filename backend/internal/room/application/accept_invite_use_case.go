package application

import (
	"backend/internal/room/domain"
	"backend/internal/shared/permissions"
	"context"
	"strings"
)

type AcceptInviteDTO struct {
	Code string `uri:"code" validate:"required"`
}

type AcceptInviteUseCase struct {
	inviteRepo     domain.InviteRepository
	roomRepo       domain.RoomRepository
	roomMemberRepo domain.RoomMemberRepository
	subscriber     domain.RealtimeSubscriber
	bans           domain.BanChecker
}

func NewAcceptInviteUseCase(inviteRepo domain.InviteRepository, roomRepo domain.RoomRepository, roomMemberRepo domain.RoomMemberRepository, subscriber domain.RealtimeSubscriber, bans domain.BanChecker) *AcceptInviteUseCase {
	return &AcceptInviteUseCase{inviteRepo: inviteRepo, roomRepo: roomRepo, roomMemberRepo: roomMemberRepo, subscriber: subscriber, bans: bans}
}

func (uc *AcceptInviteUseCase) Execute(ctx context.Context, dto AcceptInviteDTO, userID int) (*domain.Invite, error) {
	// Fetch the invite by code
	invite, err := uc.inviteRepo.GetByCode(ctx, dto.Code)
	if err != nil {
		return nil, err
	}

	// An invite addressed to an email only works for that person. Not-found hides the restriction.
	if invite.Email != "" {
		user, err := uc.roomRepo.FindActiveUser(ctx, userID)
		if err != nil {
			return nil, err
		}
		if user == nil || !strings.EqualFold(strings.TrimSpace(user.Email), strings.TrimSpace(invite.Email)) {
			return nil, domain.ErrInviteNotFound
		}
	}

	// The code dies with the creator's power to invite (left the room, or was demoted).
	creator, err := uc.roomMemberRepo.FindMember(ctx, invite.RoomID, invite.InvitedByID)
	if err != nil {
		return nil, err
	}
	if creator == nil || !(permissions.Actor{Room: permissions.RoomRole(creator.Role), IsMember: true}).CanInvite() {
		return nil, domain.ErrInviteNotFound
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
