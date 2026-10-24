package usecases

import (
	"context"

	"backend/internal/users/domain"
)

// PresenceChecker tells who has an open connection.
type PresenceChecker interface {
	OnlineUsers(ctx context.Context, userIDs []int) (map[int]bool, error)
	// OnlineUserIDs lists up to max online user ids.
	OnlineUserIDs(ctx context.Context, max int) ([]int, error)
}

const (
	directoryDefaultLimit = 50
	directoryMaxLimit     = 100
	// directoryOnlineWindow bounds how many online ids take part in an online-first search.
	directoryOnlineWindow = 200
)

// ListDirectoryUseCase lets anyone find other people, online ones first, so they can start a
// direct conversation without sharing a room.
type ListDirectoryUseCase struct {
	repo     domain.UsersRepository
	presence PresenceChecker
}

func NewListDirectoryUseCase(repo domain.UsersRepository, presence PresenceChecker) *ListDirectoryUseCase {
	return &ListDirectoryUseCase{repo: repo, presence: presence}
}

// Execute returns up to limit people matching q: first the online ones (alphabetical), then
// the rest (alphabetical). Online people are looked up by id in the database, so they surface
// even when they would fall beyond the first alphabetical page.
func (uc *ListDirectoryUseCase) Execute(ctx context.Context, selfID int, q string, limit int) ([]*domain.DirectoryEntry, error) {
	if limit <= 0 {
		limit = directoryDefaultLimit
	}
	if limit > directoryMaxLimit {
		limit = directoryMaxLimit
	}
	if uc.presence == nil {
		return uc.repo.ListDirectory(ctx, selfID, q, limit)
	}
	onlineIDs, err := uc.presence.OnlineUserIDs(ctx, directoryOnlineWindow)
	if err != nil {
		// presence is a nicety: list the people even if it is unavailable
		return uc.repo.ListDirectory(ctx, selfID, q, limit)
	}
	result := []*domain.DirectoryEntry{}
	taken := map[int]bool{}
	if len(onlineIDs) > 0 {
		online, err := uc.repo.ListDirectoryByIDs(ctx, selfID, q, onlineIDs, limit)
		if err != nil {
			return nil, err
		}
		for _, e := range online {
			e.Online = true
			taken[e.ID] = true
			result = append(result, e)
		}
	}
	if len(result) >= limit {
		return result, nil
	}
	// Fill with the alphabetical rest; the window is widened by the people already taken.
	rest, err := uc.repo.ListDirectory(ctx, selfID, q, limit+len(result))
	if err != nil {
		return nil, err
	}
	var restIDs []int
	for _, e := range rest {
		if !taken[e.ID] {
			restIDs = append(restIDs, e.ID)
		}
	}
	// Online people beyond the id window are still flagged correctly.
	var stillOnline map[int]bool
	if len(restIDs) > 0 {
		stillOnline, _ = uc.presence.OnlineUsers(ctx, restIDs)
	}
	var offline []*domain.DirectoryEntry
	for _, e := range rest {
		if taken[e.ID] {
			continue
		}
		if stillOnline[e.ID] {
			e.Online = true
			result = append(result, e)
		} else {
			offline = append(offline, e)
		}
	}
	result = append(result, offline...)
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}
