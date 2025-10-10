package application

import (
	"backend/internal/room/domain"
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrRoomIsPrivate = errors.New("cannot join a private room without an invitation")
	ErrAlreadyMember = errors.New("user is already a member of this room")
)

type JoinRoomUseCase struct {
	roomRepo   domain.RoomRepository
	memberRepo domain.RoomMemberRepository
}

func NewJoinRoomUseCase(roomRepo domain.RoomRepository, memberRepo domain.RoomMemberRepository) *JoinRoomUseCase {
	return &JoinRoomUseCase{roomRepo: roomRepo, memberRepo: memberRepo}
}

func (uc *JoinRoomUseCase) Execute(ctx context.Context, roomID uuid.UUID, userID int) error {
	room, err := uc.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return errors.New("room not found")
	}
	if room.IsPrivate {
		return ErrRoomIsPrivate
	}
	existingMember, err := uc.memberRepo.FindMember(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if existingMember != nil {
		return ErrAlreadyMember
	}
	newMember := &domain.RoomMember{RoomID: roomID, UserID: userID, Role: domain.RoleMember}
	return uc.memberRepo.AddMember(ctx, newMember)
}
