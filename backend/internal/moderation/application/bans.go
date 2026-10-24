package application

import (
	"context"
	"strings"
	"time"

	"backend/internal/moderation/domain"
	shared_domain "backend/internal/shared/domain"

	"github.com/google/uuid"
)

// maxBanDuration keeps typos from creating bans that last for centuries; leave the duration
// empty for a permanent ban.
const maxBanDuration = 10 * 365 * 24 * time.Hour

// BansUseCase bans and unbans users from a room.
type BansUseCase struct {
	repo     domain.Repository
	realtime domain.Realtime
	now      func() time.Time
}

func NewBansUseCase(repo domain.Repository, realtime domain.Realtime) *BansUseCase {
	return &BansUseCase{repo: repo, realtime: realtime, now: time.Now}
}

// Ban removes the user from the room and stops them coming back. duration is a Go duration
// ("24h", "168h"); empty means permanent. Moderators need to outrank the target.
func (uc *BansUseCase) Ban(ctx context.Context, roomID uuid.UUID, actorID, targetID int, reason, duration string) error {
	if actorID == targetID {
		return shared_domain.ErrInvalidInput
	}
	var expires *time.Time
	if duration = strings.TrimSpace(duration); duration != "" {
		d, err := time.ParseDuration(duration)
		if err != nil || d <= 0 || d > maxBanDuration {
			return shared_domain.ErrInvalidInput
		}
		t := uc.now().Add(d)
		expires = &t
	}

	actor, err := uc.repo.ActorIn(ctx, roomID, actorID)
	if err != nil {
		return err
	}
	target, err := uc.repo.ActorIn(ctx, roomID, targetID)
	if err != nil {
		return err
	}
	if target.IsMember && target.Room == "owner" {
		return domain.ErrCannotTargetOwner
	}
	if !actor.CanModerate(target.Rank()) {
		return shared_domain.ErrForbidden
	}

	if err := uc.repo.UpsertBan(ctx, domain.Ban{UserID: targetID, RoomID: roomID, BannedBy: &actorID, Reason: reason, ExpiresAt: expires}); err != nil {
		return err
	}
	if target.IsMember {
		if err := uc.repo.RemoveMember(ctx, roomID, targetID); err != nil {
			return err
		}
		if uc.realtime != nil {
			uc.realtime.Broadcast("member_removed", map[string]interface{}{"room_id": roomID, "user_id": targetID}, roomID.String())
			uc.realtime.UnsubscribeUser(targetID, roomID.String())
		}
	}

	details := map[string]interface{}{"reason": reason}
	if expires != nil {
		details["expires_at"] = expires
	}
	if err := uc.repo.AddAudit(ctx, domain.AuditEntry{
		ActorID: &actorID, Action: domain.ActionMemberBanned, RoomID: &roomID, TargetUserID: &targetID, Details: details,
	}); err != nil {
		logAuditFailure(domain.ActionMemberBanned, err)
	}
	return nil
}

// Unban lifts a ban. Moderators and above may do it.
func (uc *BansUseCase) Unban(ctx context.Context, roomID uuid.UUID, actorID, targetID int) error {
	actor, err := uc.repo.ActorIn(ctx, roomID, actorID)
	if err != nil {
		return err
	}
	if !actor.CanResolveReports() { // same bar as handling reports: moderator and above
		return shared_domain.ErrForbidden
	}
	removed, err := uc.repo.DeleteBan(ctx, roomID, targetID)
	if err != nil {
		return err
	}
	if !removed {
		return shared_domain.ErrNotFound
	}
	if err := uc.repo.AddAudit(ctx, domain.AuditEntry{
		ActorID: &actorID, Action: domain.ActionMemberUnbanned, RoomID: &roomID, TargetUserID: &targetID, Details: map[string]interface{}{},
	}); err != nil {
		logAuditFailure(domain.ActionMemberUnbanned, err)
	}
	return nil
}

// List shows the room's active bans to moderators and above.
func (uc *BansUseCase) List(ctx context.Context, roomID uuid.UUID, userID int) ([]*domain.Ban, error) {
	actor, err := uc.repo.ActorIn(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if !actor.CanResolveReports() {
		return nil, shared_domain.ErrForbidden
	}
	return uc.repo.ListBans(ctx, roomID)
}
