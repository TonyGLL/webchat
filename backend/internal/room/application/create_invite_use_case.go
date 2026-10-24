package application

import (
	"backend/internal/room/domain"
	shared_domain "backend/internal/shared/domain"
	"backend/internal/shared/permissions"
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

const inviteCodeBytes = 8

// An invite code lives between a minute and 30 days: a "permanent" code is a standing hole.
const (
	MinInviteTTL = time.Minute
	MaxInviteTTL = 30 * 24 * time.Hour
)

type CreateInviteDTO struct {
	RoomID    string `uri:"room_id" validate:"required,uuid"`
	ExpiresIn string `json:"expires_in" validate:"required"` // a Go duration between 1m and 720h, e.g. "24h"
	Email     string `json:"email" validate:"omitempty,email"`
}

type CreateInviteUseCase struct {
	inviteRepo domain.InviteRepository
	roomRepo   domain.RoomRepository
	memberRepo domain.RoomMemberRepository
}

func NewCreateInviteUseCase(inviteRepo domain.InviteRepository, roomRepo domain.RoomRepository, memberRepo domain.RoomMemberRepository) *CreateInviteUseCase {
	return &CreateInviteUseCase{inviteRepo: inviteRepo, roomRepo: roomRepo, memberRepo: memberRepo}
}

func (uc *CreateInviteUseCase) Execute(ctx context.Context, dto CreateInviteDTO, invitedByID int) (*domain.Invite, error) {
	roomID, err := uuid.Parse(dto.RoomID)
	if err != nil {
		return nil, shared_domain.ErrInvalidInput
	}

	// Only the owner or an admin of the room can invite people.
	member, err := uc.memberRepo.FindMember(ctx, roomID, invitedByID)
	if err != nil {
		return nil, err
	}
	if member == nil || !(permissions.Actor{Room: permissions.RoomRole(member.Role), IsMember: true}).CanInvite() {
		return nil, shared_domain.ErrForbidden
	}

	room, err := uc.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room != nil && room.IsDirect {
		return nil, shared_domain.ErrForbidden // a conversation between two people cannot grow
	}

	duration, err := time.ParseDuration(dto.ExpiresIn)
	if err != nil || duration < MinInviteTTL || duration > MaxInviteTTL {
		return nil, shared_domain.ErrInvalidInput
	}

	code, err := generateInviteCode(inviteCodeBytes)
	if err != nil {
		return nil, err
	}

	invite := &domain.Invite{
		RoomID:      roomID,
		Code:        code,
		ExpiresAt:   time.Now().Add(duration),
		InvitedByID: invitedByID,
		Email:       dto.Email,
	}

	return uc.inviteRepo.Create(ctx, invite)
}

func generateInviteCode(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
