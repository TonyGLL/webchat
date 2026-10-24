package usecases

import (
	"context"

	"backend/internal/users/domain"
)

// PresenceChecker tells which of the given users have an open connection.
type PresenceChecker interface {
	OnlineUsers(ctx context.Context, userIDs []int) (map[int]bool, error)
}

const (
	directoryDefaultLimit = 50
	directoryMaxLimit     = 100
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

func (uc *ListDirectoryUseCase) Execute(ctx context.Context, selfID int, q string, limit int) ([]*domain.DirectoryEntry, error) {
	if limit <= 0 {
		limit = directoryDefaultLimit
	}
	if limit > directoryMaxLimit {
		limit = directoryMaxLimit
	}
	entries, err := uc.repo.ListDirectory(ctx, selfID, q, limit)
	if err != nil || uc.presence == nil || len(entries) == 0 {
		return entries, err
	}
	ids := make([]int, len(entries))
	for i, e := range entries {
		ids[i] = e.ID
	}
	online, err := uc.presence.OnlineUsers(ctx, ids)
	if err != nil {
		return entries, nil // presence is a nicety: list the people even if it is unavailable
	}
	// Stable partition: online first, keeping the alphabetical order inside each group.
	sorted := make([]*domain.DirectoryEntry, 0, len(entries))
	for _, e := range entries {
		e.Online = online[e.ID]
		if e.Online {
			sorted = append(sorted, e)
		}
	}
	for _, e := range entries {
		if !e.Online {
			sorted = append(sorted, e)
		}
	}
	return sorted, nil
}
