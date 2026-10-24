package application

import (
	"context"
	"errors"
	"testing"

	"backend/internal/moderation/domain"
	shared_domain "backend/internal/shared/domain"
	"backend/internal/shared/permissions"

	"github.com/google/uuid"
)

type fakeRepo struct {
	actors map[int]permissions.Actor
	audit  []domain.AuditEntry
}

func (f *fakeRepo) ActorIn(_ context.Context, _ uuid.UUID, id int) (permissions.Actor, error) {
	return f.actors[id], nil
}
func (f *fakeRepo) SetMemberRole(_ context.Context, _ uuid.UUID, id int, role permissions.RoomRole) error {
	a := f.actors[id]
	a.Room = role
	f.actors[id] = a
	return nil
}
func (f *fakeRepo) RemoveMember(_ context.Context, _ uuid.UUID, id int) error {
	a := f.actors[id]
	a.IsMember = false
	f.actors[id] = a
	return nil
}
func (f *fakeRepo) AddAudit(_ context.Context, e domain.AuditEntry) error {
	f.audit = append(f.audit, e)
	return nil
}
func (f *fakeRepo) RecordMessageDeletedByModerator(context.Context, int, uuid.UUID, uuid.UUID, int) error {
	return nil
}
func (f *fakeRepo) ListAudit(context.Context, uuid.UUID, int) ([]*domain.AuditEntry, error) {
	return nil, nil
}

type spyRealtime struct{ events, unsubscribed []string }

func (s *spyRealtime) Broadcast(t string, _ interface{}, _ string) { s.events = append(s.events, t) }
func (s *spyRealtime) UnsubscribeUser(_ int, room string) {
	s.unsubscribed = append(s.unsubscribed, room)
}

const (
	ownerID, adminID, modID, memberID, otherMemberID, strangerID, globalAdminID, otherModID = 1, 2, 3, 4, 5, 6, 7, 8
)

func setup() (*ManageMembersUseCase, *fakeRepo, *spyRealtime, uuid.UUID) {
	repo := &fakeRepo{actors: map[int]permissions.Actor{
		ownerID:       {Room: permissions.RoomOwner, IsMember: true, Global: permissions.GlobalUser},
		adminID:       {Room: permissions.RoomAdmin, IsMember: true, Global: permissions.GlobalUser},
		modID:         {Room: permissions.RoomModerator, IsMember: true, Global: permissions.GlobalUser},
		memberID:      {Room: permissions.RoomMember, IsMember: true, Global: permissions.GlobalUser},
		otherMemberID: {Room: permissions.RoomMember, IsMember: true, Global: permissions.GlobalUser},
		otherModID:    {Room: permissions.RoomModerator, IsMember: true, Global: permissions.GlobalUser},
		strangerID:    {Global: permissions.GlobalUser},
		globalAdminID: {Global: permissions.GlobalAdmin},
	}}
	rt := &spyRealtime{}
	return NewManageMembersUseCase(repo, rt), repo, rt, uuid.New()
}

func TestChangeRoleFollowsThePermissionMatrix(t *testing.T) {
	uc, repo, rt, room := setup()
	ctx := context.Background()

	cases := []struct {
		name          string
		actor, target int
		role          string
		wantErr       error
	}{
		{"member cannot promote", memberID, otherMemberID, "moderator", shared_domain.ErrForbidden},
		{"moderator cannot promote", modID, memberID, "moderator", shared_domain.ErrForbidden},
		{"admin cannot make an admin", adminID, memberID, "admin", shared_domain.ErrForbidden},
		{"nobody can change the owner", globalAdminID, ownerID, "member", domain.ErrCannotTargetOwner},
		{"cannot change yourself", adminID, adminID, "member", shared_domain.ErrInvalidInput},
		{"unknown role", ownerID, memberID, "wizard", shared_domain.ErrInvalidInput},
		{"owner role cannot be assigned", ownerID, memberID, "owner", shared_domain.ErrForbidden},
		{"target not in the room", ownerID, strangerID, "moderator", domain.ErrNotMember},
		{"admin makes a moderator", adminID, memberID, "moderator", nil},
		{"owner makes an admin", ownerID, otherMemberID, "admin", nil},
		{"global admin demotes an admin", globalAdminID, adminID, "member", nil},
	}
	for _, c := range cases {
		err := uc.ChangeRole(ctx, room, c.actor, c.target, c.role)
		if !errors.Is(err, c.wantErr) {
			t.Errorf("%s: err = %v, want %v", c.name, err, c.wantErr)
		}
	}

	if repo.actors[memberID].Room != permissions.RoomModerator || repo.actors[otherMemberID].Room != permissions.RoomAdmin {
		t.Error("successful changes should be stored")
	}
	if repo.actors[ownerID].Room != permissions.RoomOwner {
		t.Error("the owner must never change")
	}
	if len(repo.audit) != 3 || repo.audit[0].Action != domain.ActionRoleChanged {
		t.Errorf("each successful change must be audited, got %d entries", len(repo.audit))
	}
	if len(rt.events) != 3 {
		t.Errorf("each change must be broadcast, got %v", rt.events)
	}
}

func TestKickRequiresOutranking(t *testing.T) {
	uc, repo, rt, room := setup()
	ctx := context.Background()

	for name, c := range map[string][2]int{
		"member kicks member":       {memberID, otherMemberID},
		"moderator kicks moderator": {modID, otherModID},
		"moderator kicks admin":     {modID, adminID},
		"admin kicks owner":         {adminID, ownerID},
	} {
		if err := uc.Kick(ctx, room, c[0], c[1], ""); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
	if err := uc.Kick(ctx, room, adminID, adminID, ""); !errors.Is(err, shared_domain.ErrInvalidInput) {
		t.Errorf("kicking yourself: err = %v", err)
	}

	if err := uc.Kick(ctx, room, modID, memberID, "spam"); err != nil {
		t.Fatalf("moderator kicks member: %v", err)
	}
	if repo.actors[memberID].IsMember {
		t.Error("kicked member should be removed")
	}
	if len(rt.unsubscribed) != 1 || len(repo.audit) != 1 || repo.audit[0].Action != domain.ActionMemberKicked {
		t.Errorf("kick must cut realtime and be audited: %v / %v", rt.unsubscribed, repo.audit)
	}
}

func TestAuditIsForAdminsOnly(t *testing.T) {
	_, repo, _, room := setup()
	uc := NewListAuditUseCase(repo)
	ctx := context.Background()

	for _, id := range []int{memberID, modID, strangerID} {
		if _, err := uc.Execute(ctx, room, id, 10); !errors.Is(err, shared_domain.ErrForbidden) {
			t.Errorf("user %d: err = %v, want ErrForbidden", id, err)
		}
	}
	for _, id := range []int{adminID, ownerID, globalAdminID} {
		if _, err := uc.Execute(ctx, room, id, 10); err != nil {
			t.Errorf("user %d: unexpected err %v", id, err)
		}
	}
}
