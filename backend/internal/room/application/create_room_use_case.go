package application

import (
	"backend/internal/room/domain"
	"context"
)

type CreateRoomDTO struct {
	Name      string  `json:"name" validate:"required,min=3,max=150"`
	Topic     *string `json:"topic" validate:"max=255"`
	IsPrivate bool    `json:"is_private"`
}

type CreateRoomUseCase struct {
	roomRepo   domain.RoomRepository
	memberRepo domain.RoomMemberRepository
}

func NewCreateRoomUseCase(roomRepo domain.RoomRepository, memberRepo domain.RoomMemberRepository) *CreateRoomUseCase {
	return &CreateRoomUseCase{roomRepo: roomRepo, memberRepo: memberRepo}
}

func (uc *CreateRoomUseCase) Execute(ctx context.Context, input CreateRoomDTO, ownerID int) (*domain.Room, error) {
	var createdRoom *domain.Room
	room := &domain.Room{OwnerID: ownerID, Name: input.Name, Topic: input.Topic, IsPrivate: input.IsPrivate}
	var err error
	createdRoom, err = uc.roomRepo.Create(ctx, room)
	if err != nil {
		return nil, err
	}

	member := &domain.RoomMember{RoomID: createdRoom.ID, UserID: ownerID, Role: domain.RoleOwner}
	err = uc.memberRepo.AddMember(ctx, member)
	if err != nil {
		return nil, err
	}

	return createdRoom, nil
}
