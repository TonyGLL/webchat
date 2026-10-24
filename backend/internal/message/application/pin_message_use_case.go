package application

import (
	"context"

	"backend/internal/message/domain"
	shared_domain "backend/internal/shared/domain"

	"github.com/google/uuid"
)

// maxPinned bounds how many messages a room can have pinned at once.
const maxPinned = 20

// PinMessageUseCase pins and unpins messages. Only room owners and admins may do it.
type PinMessageUseCase struct {
	messageRepo   domain.MessageRepository
	wsBroadcaster domain.WebsocketBroadcaster
}

func NewPinMessageUseCase(messageRepo domain.MessageRepository, wsBroadcaster domain.WebsocketBroadcaster) *PinMessageUseCase {
	return &PinMessageUseCase{messageRepo: messageRepo, wsBroadcaster: wsBroadcaster}
}

func (uc *PinMessageUseCase) Execute(ctx context.Context, messageID uuid.UUID, userID int, pinned bool) error {
	msg, err := uc.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return err
	}
	if msg.DeletedAt != nil {
		return domain.ErrMessageNotFound
	}

	actor, err := uc.messageRepo.ActorIn(ctx, msg.RoomID, userID)
	if err != nil {
		return err
	}
	if !actor.CanPin() {
		return shared_domain.ErrForbidden
	}

	if pinned {
		current, err := uc.messageRepo.FindPinned(ctx, msg.RoomID)
		if err != nil {
			return err
		}
		if len(current) >= maxPinned {
			return shared_domain.ErrInvalidInput
		}
	}
	if err := uc.messageRepo.SetPinned(ctx, messageID, pinned); err != nil {
		return err
	}
	if uc.wsBroadcaster != nil {
		uc.wsBroadcaster.Broadcast("pin_message", map[string]interface{}{"message_id": messageID, "pinned": pinned}, msg.RoomID.String())
	}
	return nil
}

// ListPinnedUseCase returns a room's pinned messages, newest pin first.
type ListPinnedUseCase struct {
	messageRepo domain.MessageRepository
}

func NewListPinnedUseCase(messageRepo domain.MessageRepository) *ListPinnedUseCase {
	return &ListPinnedUseCase{messageRepo: messageRepo}
}

func (uc *ListPinnedUseCase) Execute(ctx context.Context, roomIDStr string, userID int) ([]*domain.Message, error) {
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		return nil, shared_domain.ErrInvalidInput
	}
	if err := requireRoomMember(ctx, uc.messageRepo, roomID, userID); err != nil {
		return nil, err
	}
	return uc.messageRepo.FindPinned(ctx, roomID)
}
