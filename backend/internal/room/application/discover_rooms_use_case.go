package application

import (
	"context"

	"backend/internal/room/domain"

	"github.com/google/uuid"
)

const (
	DefaultDiscoverLimit = 30
	MaxDiscoverLimit     = 100
)

// DiscoverRoomsUseCase lists rooms so people can find where to talk: public rooms can be
// joined straight away, private ones are listed but need an invitation to read or write.
// It is paginated by (name, id): pass the id of the last room received as `after`.
type DiscoverRoomsUseCase struct {
	roomRepo domain.RoomRepository
}

func NewDiscoverRoomsUseCase(roomRepo domain.RoomRepository) *DiscoverRoomsUseCase {
	return &DiscoverRoomsUseCase{roomRepo: roomRepo}
}

// Execute returns up to limit rooms (default 30, at most 100) after the given room, if any.
func (uc *DiscoverRoomsUseCase) Execute(ctx context.Context, userID, limit int, after *uuid.UUID) ([]*domain.RoomSummary, error) {
	return uc.roomRepo.ListDiscoverable(ctx, userID, clampLimit(limit, DefaultDiscoverLimit, MaxDiscoverLimit), after)
}

// clampLimit applies the default to a missing (<= 0) limit and the maximum to a large one.
func clampLimit(limit, def, max int) int {
	switch {
	case limit <= 0:
		return def
	case limit > max:
		return max
	}
	return limit
}
