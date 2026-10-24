package application

import (
	"context"
	"log/slog"

	"backend/internal/moderation/domain"
	shared_domain "backend/internal/shared/domain"
	"backend/internal/shared/permissions"

	"github.com/google/uuid"
)

func logAuditFailure(action string, err error) {
	slog.Error("could not write audit log", "action", action, "error", err)
}

const (
	defaultAuditLimit = 50
	maxAuditLimit     = 200
)

// ListAuditUseCase shows a room's audit log to admins and above.
type ListAuditUseCase struct {
	repo domain.Repository
}

func NewListAuditUseCase(repo domain.Repository) *ListAuditUseCase {
	return &ListAuditUseCase{repo: repo}
}

func (uc *ListAuditUseCase) Execute(ctx context.Context, roomID uuid.UUID, userID, limit int) ([]*domain.AuditEntry, error) {
	actor, err := uc.repo.ActorIn(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if actor.Rank() < permissions.RankAdmin {
		return nil, shared_domain.ErrForbidden
	}
	if limit <= 0 {
		limit = defaultAuditLimit
	} else if limit > maxAuditLimit {
		limit = maxAuditLimit
	}
	return uc.repo.ListAudit(ctx, roomID, limit)
}
