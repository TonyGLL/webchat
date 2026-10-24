package application

import (
	"context"

	"backend/internal/room/domain"
	shared_domain "backend/internal/shared/domain"
)

// OpenDirectUseCase opens (creating if needed) the one-to-one conversation with another user.
type OpenDirectUseCase struct {
	roomRepo   domain.RoomRepository
	memberRepo domain.RoomMemberRepository
	subscriber domain.RealtimeSubscriber
}

func NewOpenDirectUseCase(roomRepo domain.RoomRepository, memberRepo domain.RoomMemberRepository, subscriber domain.RealtimeSubscriber) *OpenDirectUseCase {
	return &OpenDirectUseCase{roomRepo: roomRepo, memberRepo: memberRepo, subscriber: subscriber}
}

func (uc *OpenDirectUseCase) Execute(ctx context.Context, userID, otherID int) (*domain.Room, error) {
	if userID == otherID {
		return nil, shared_domain.ErrInvalidInput
	}
	// A missing or deactivated person cannot be messaged (same answer for both).
	other, err := uc.roomRepo.FindActiveUser(ctx, otherID)
	if err != nil {
		return nil, err
	}
	if other == nil {
		return nil, domain.ErrUserNotFound
	}
	room, err := uc.roomRepo.FindOrCreateDirect(ctx, userID, otherID)
	if err != nil {
		return nil, err
	}

	// If this user left the conversation earlier, bring them back into it.
	member, err := uc.memberRepo.FindMember(ctx, room.ID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		if err := uc.memberRepo.AddMember(ctx, &domain.RoomMember{RoomID: room.ID, UserID: userID, Role: domain.RoleMember}); err != nil {
			return nil, err
		}
	}

	if uc.subscriber != nil {
		roomID := room.ID.String()
		uc.subscriber.SubscribeUser(userID, roomID)
		uc.subscriber.SubscribeUser(otherID, roomID) // so the other person receives it live too
	}
	return room, nil
}
