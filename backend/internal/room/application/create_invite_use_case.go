package application

import (
	"backend/internal/room/domain"
	shared_domain "backend/internal/shared/domain"
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

const inviteCodeBytes = 8

type CreateInviteDTO struct {
	RoomID    string `uri:"room_id" validate:"required,uuid"`
	ExpiresIn string `json:"expires_in" validate:"required"` // e.g., "24h", "7d"
	Email     string `json:"email" validate:"omitempty,email"`
}

type CreateInviteUseCase struct {
	inviteRepo domain.InviteRepository
	roomRepo   domain.RoomRepository
}

func NewCreateInviteUseCase(inviteRepo domain.InviteRepository, roomRepo domain.RoomRepository) *CreateInviteUseCase {
	return &CreateInviteUseCase{inviteRepo: inviteRepo, roomRepo: roomRepo}
}

func (uc *CreateInviteUseCase) Execute(ctx context.Context, dto CreateInviteDTO, invitedByID int) (*domain.Invite, error) {
	roomID, err := uuid.Parse(dto.RoomID)
	if err != nil {
		return nil, shared_domain.ErrInvalidInput
	}

	// Check if the user is a member of the room
	// Note: We might need a RoomMemberRepository for this check.
	// Assuming for now that the room repository can provide this info or the check is done in the controller.

	duration, err := time.ParseDuration(dto.ExpiresIn)
	if err != nil {
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
