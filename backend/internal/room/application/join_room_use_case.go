package application

import (
	"backend/internal/room/domain"
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrRoomIsPrivate  = errors.New("cannot join a private room without an invitation")
	ErrAlreadyMember  = errors.New("user is already a member of this room")
	ErrBannedFromRoom = errors.New("you are banned from this room")
	ErrRoomNotFound   = errors.New("room not found")
)

type JoinRoomUseCase struct {
	roomRepo   domain.RoomRepository
	memberRepo domain.RoomMemberRepository
	subscriber domain.RealtimeSubscriber
	bans       domain.BanChecker
}

func NewJoinRoomUseCase(roomRepo domain.RoomRepository, memberRepo domain.RoomMemberRepository, subscriber domain.RealtimeSubscriber, bans domain.BanChecker) *JoinRoomUseCase {
	return &JoinRoomUseCase{roomRepo: roomRepo, memberRepo: memberRepo, subscriber: subscriber, bans: bans}
}

func (uc *JoinRoomUseCase) Execute(ctx context.Context, roomID uuid.UUID, userID int) error {
	room, err := uc.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return ErrRoomNotFound
	}
	if room.IsPrivate {
		return ErrRoomIsPrivate
	}
	if uc.bans != nil {
		banned, err := uc.bans.IsBanned(ctx, roomID, userID)
		if err != nil {
			return err
		}
		if banned {
			return ErrBannedFromRoom
		}
	}
	existingMember, err := uc.memberRepo.FindMember(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if existingMember != nil {
		return ErrAlreadyMember
	}
	newMember := &domain.RoomMember{RoomID: roomID, UserID: userID, Role: domain.RoleMember}
	if err := uc.memberRepo.AddMember(ctx, newMember); err != nil {
		return err
	}
	if uc.subscriber != nil {
		uc.subscriber.SubscribeUser(userID, roomID.String())
	}
	return nil
}
