package application

import (
	"context"

	"backend/internal/moderation/domain"
	shared_domain "backend/internal/shared/domain"
	"backend/internal/shared/permissions"

	"github.com/google/uuid"
)

// ManageMembersUseCase changes member roles and removes members.
type ManageMembersUseCase struct {
	repo     domain.Repository
	realtime domain.Realtime
}

func NewManageMembersUseCase(repo domain.Repository, realtime domain.Realtime) *ManageMembersUseCase {
	return &ManageMembersUseCase{repo: repo, realtime: realtime}
}

// loadTarget returns the target as an Actor, failing if they are not a member.
func (uc *ManageMembersUseCase) loadTarget(ctx context.Context, roomID uuid.UUID, targetID int) (permissions.Actor, error) {
	target, err := uc.repo.ActorIn(ctx, roomID, targetID)
	if err != nil {
		return permissions.Actor{}, err
	}
	if !target.IsMember {
		return permissions.Actor{}, domain.ErrNotMember
	}
	if target.Room == permissions.RoomOwner {
		return permissions.Actor{}, domain.ErrCannotTargetOwner
	}
	return target, nil
}

// ChangeRole sets a member's room role (member, moderator or admin).
func (uc *ManageMembersUseCase) ChangeRole(ctx context.Context, roomID uuid.UUID, actorID, targetID int, newRole string) error {
	if actorID == targetID || !permissions.ValidRoomRole(newRole) {
		return shared_domain.ErrInvalidInput
	}
	role := permissions.RoomRole(newRole)

	actor, err := uc.repo.ActorIn(ctx, roomID, actorID)
	if err != nil {
		return err
	}
	target, err := uc.loadTarget(ctx, roomID, targetID)
	if err != nil {
		return err
	}
	if !actor.CanAssignRole(target.Rank(), role) {
		return shared_domain.ErrForbidden
	}
	if target.Room == role {
		return nil // nothing to change
	}

	if err := uc.repo.SetMemberRole(ctx, roomID, targetID, role); err != nil {
		return err
	}
	uc.audit(ctx, domain.AuditEntry{
		ActorID: &actorID, Action: domain.ActionRoleChanged, RoomID: &roomID, TargetUserID: &targetID,
		Details: map[string]interface{}{"from": target.Room, "to": role},
	})
	if uc.realtime != nil {
		uc.realtime.Broadcast("role_changed", map[string]interface{}{"room_id": roomID, "user_id": targetID, "role": role}, roomID.String())
	}
	return nil
}

// Kick removes a member from the room. They can come back if the room lets them (see bans).
func (uc *ManageMembersUseCase) Kick(ctx context.Context, roomID uuid.UUID, actorID, targetID int, reason string) error {
	if actorID == targetID {
		return shared_domain.ErrInvalidInput
	}
	actor, err := uc.repo.ActorIn(ctx, roomID, actorID)
	if err != nil {
		return err
	}
	target, err := uc.loadTarget(ctx, roomID, targetID)
	if err != nil {
		return err
	}
	if !actor.CanModerate(target.Rank()) {
		return shared_domain.ErrForbidden
	}

	if err := uc.repo.RemoveMember(ctx, roomID, targetID); err != nil {
		return err
	}
	uc.audit(ctx, domain.AuditEntry{
		ActorID: &actorID, Action: domain.ActionMemberKicked, RoomID: &roomID, TargetUserID: &targetID,
		Details: map[string]interface{}{"reason": reason},
	})
	if uc.realtime != nil {
		uc.realtime.Broadcast("member_removed", map[string]interface{}{"room_id": roomID, "user_id": targetID}, roomID.String())
		uc.realtime.UnsubscribeUser(targetID, roomID.String())
	}
	return nil
}

// audit records an action; a failure to write the log is reported but never undoes the action.
func (uc *ManageMembersUseCase) audit(ctx context.Context, entry domain.AuditEntry) {
	if err := uc.repo.AddAudit(ctx, entry); err != nil {
		logAuditFailure(entry.Action, err)
	}
}
