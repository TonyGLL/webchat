package application

import (
	"context"

	"backend/internal/room/domain"
	shared_domain "backend/internal/shared/domain"
	"backend/internal/shared/permissions"

	"github.com/google/uuid"
)

// InviteUserUseCase sends an invitation to one person; only members who can invite (admins and
// above) may do it, never in a direct conversation.
type InviteUserUseCase struct {
	invitations domain.InvitationRepository
	roomRepo    domain.RoomRepository
	memberRepo  domain.RoomMemberRepository
	bans        domain.BanChecker
}

func NewInviteUserUseCase(invitations domain.InvitationRepository, roomRepo domain.RoomRepository, memberRepo domain.RoomMemberRepository, bans domain.BanChecker) *InviteUserUseCase {
	return &InviteUserUseCase{invitations: invitations, roomRepo: roomRepo, memberRepo: memberRepo, bans: bans}
}

func (uc *InviteUserUseCase) Execute(ctx context.Context, roomID uuid.UUID, inviterID, inviteeID int) (*domain.Invitation, error) {
	if inviterID == inviteeID {
		return nil, shared_domain.ErrInvalidInput
	}
	inviter, err := uc.memberRepo.FindMember(ctx, roomID, inviterID)
	if err != nil {
		return nil, err
	}
	if inviter == nil || !(permissions.Actor{Room: permissions.RoomRole(inviter.Role), IsMember: true}).CanInvite() {
		return nil, shared_domain.ErrForbidden
	}
	room, err := uc.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, ErrRoomNotFound
	}
	if room.IsDirect {
		return nil, shared_domain.ErrForbidden
	}
	existing, err := uc.memberRepo.FindMember(ctx, roomID, inviteeID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyMember
	}
	if uc.bans != nil {
		banned, err := uc.bans.IsBanned(ctx, roomID, inviteeID)
		if err != nil {
			return nil, err
		}
		if banned {
			return nil, ErrBannedFromRoom
		}
	}
	return uc.invitations.Create(ctx, roomID, inviterID, inviteeID)
}

// ListInvitationsUseCase returns the invitations waiting for a user's answer.
type ListInvitationsUseCase struct {
	invitations domain.InvitationRepository
}

func NewListInvitationsUseCase(invitations domain.InvitationRepository) *ListInvitationsUseCase {
	return &ListInvitationsUseCase{invitations: invitations}
}

func (uc *ListInvitationsUseCase) Execute(ctx context.Context, userID int) ([]*domain.Invitation, error) {
	return uc.invitations.ListPending(ctx, userID)
}

// ListRoomInvitationsUseCase shows a room's pending invitations to those who can invite.
type ListRoomInvitationsUseCase struct {
	invitations domain.InvitationRepository
	memberRepo  domain.RoomMemberRepository
}

func NewListRoomInvitationsUseCase(invitations domain.InvitationRepository, memberRepo domain.RoomMemberRepository) *ListRoomInvitationsUseCase {
	return &ListRoomInvitationsUseCase{invitations: invitations, memberRepo: memberRepo}
}

func canInviteIn(ctx context.Context, memberRepo domain.RoomMemberRepository, roomID uuid.UUID, userID int) error {
	m, err := memberRepo.FindMember(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if m == nil || !(permissions.Actor{Room: permissions.RoomRole(m.Role), IsMember: true}).CanInvite() {
		return shared_domain.ErrForbidden
	}
	return nil
}

func (uc *ListRoomInvitationsUseCase) Execute(ctx context.Context, roomID uuid.UUID, userID int) ([]*domain.Invitation, error) {
	if err := canInviteIn(ctx, uc.memberRepo, roomID, userID); err != nil {
		return nil, err
	}
	return uc.invitations.ListByRoom(ctx, roomID)
}

// CancelInvitationUseCase lets admins and above withdraw a pending invitation of their room.
type CancelInvitationUseCase struct {
	invitations domain.InvitationRepository
	memberRepo  domain.RoomMemberRepository
}

func NewCancelInvitationUseCase(invitations domain.InvitationRepository, memberRepo domain.RoomMemberRepository) *CancelInvitationUseCase {
	return &CancelInvitationUseCase{invitations: invitations, memberRepo: memberRepo}
}

func (uc *CancelInvitationUseCase) Execute(ctx context.Context, roomID, id uuid.UUID, userID int) error {
	if err := canInviteIn(ctx, uc.memberRepo, roomID, userID); err != nil {
		return err
	}
	ok, err := uc.invitations.Cancel(ctx, roomID, id)
	if err != nil {
		return err
	}
	if !ok {
		return shared_domain.ErrNotFound
	}
	return nil
}

// RespondInvitationUseCase lets the invited user accept (becoming a member, with access to the
// history and the right to write) or decline.
type RespondInvitationUseCase struct {
	invitations domain.InvitationRepository
	memberRepo  domain.RoomMemberRepository
	subscriber  domain.RealtimeSubscriber
	bans        domain.BanChecker
}

func NewRespondInvitationUseCase(invitations domain.InvitationRepository, memberRepo domain.RoomMemberRepository, subscriber domain.RealtimeSubscriber, bans domain.BanChecker) *RespondInvitationUseCase {
	return &RespondInvitationUseCase{invitations: invitations, memberRepo: memberRepo, subscriber: subscriber, bans: bans}
}

func (uc *RespondInvitationUseCase) Execute(ctx context.Context, id uuid.UUID, userID int, accept bool) error {
	inv, err := uc.invitations.FindPending(ctx, id)
	if err != nil {
		return err
	}
	if inv == nil || inv.InviteeID != userID {
		return shared_domain.ErrNotFound // do not reveal other people's invitations
	}
	if accept {
		// The invitation dies with the inviter's power to invite (left the room, or was demoted),
		// like invite codes do. Not-found hides the reason.
		inviter, err := uc.memberRepo.FindMember(ctx, inv.RoomID, inv.InviterID)
		if err != nil {
			return err
		}
		if inviter == nil || !(permissions.Actor{Room: permissions.RoomRole(inviter.Role), IsMember: true}).CanInvite() {
			return shared_domain.ErrNotFound
		}
		if uc.bans != nil {
			banned, err := uc.bans.IsBanned(ctx, inv.RoomID, userID)
			if err != nil {
				return err
			}
			if banned {
				return ErrBannedFromRoom
			}
		}
		member, err := uc.memberRepo.FindMember(ctx, inv.RoomID, userID)
		if err != nil {
			return err
		}
		if member == nil {
			if err := uc.memberRepo.AddMember(ctx, &domain.RoomMember{RoomID: inv.RoomID, UserID: userID, Role: domain.RoleMember}); err != nil {
				return err
			}
		}
		if uc.subscriber != nil {
			uc.subscriber.SubscribeUser(userID, inv.RoomID.String())
		}
	}
	_, err = uc.invitations.Resolve(ctx, id, accept)
	return err
}
