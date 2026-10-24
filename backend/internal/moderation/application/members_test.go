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
	actors   map[int]permissions.Actor
	audit    []domain.AuditEntry
	messages map[uuid.UUID]*domain.MessageRef
	reports  map[int]*domain.Report
	bans     map[int]domain.Ban
	nextID   int
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
func (f *fakeRepo) MessageRef(_ context.Context, id uuid.UUID) (*domain.MessageRef, error) {
	if m, ok := f.messages[id]; ok {
		return m, nil
	}
	return nil, domain.ErrMessageNotFound
}
func (f *fakeRepo) SoftDeleteMessage(_ context.Context, id uuid.UUID) error {
	f.messages[id].Deleted = true
	return nil
}
func (f *fakeRepo) CreateReport(_ context.Context, id uuid.UUID, reporter int, reason string) (*domain.Report, error) {
	for _, r := range f.reports {
		if r.MessageID == id && r.ReportedBy == reporter {
			return nil, domain.ErrAlreadyReported
		}
	}
	f.nextID++
	r := &domain.Report{ID: f.nextID, MessageID: id, RoomID: f.messages[id].RoomID, MessageAuthor: f.messages[id].Author, ReportedBy: reporter, Reason: reason, Status: domain.ReportPending}
	f.reports[r.ID] = r
	return r, nil
}
func (f *fakeRepo) FindReport(_ context.Context, id int) (*domain.Report, error) {
	if r, ok := f.reports[id]; ok {
		return r, nil
	}
	return nil, domain.ErrReportNotFound
}
func (f *fakeRepo) ListReports(_ context.Context, _ uuid.UUID, status string, _ int) ([]*domain.Report, error) {
	out := []*domain.Report{}
	for _, r := range f.reports {
		if r.Status == status {
			out = append(out, r)
		}
	}
	return out, nil
}
func (f *fakeRepo) SetReportStatus(_ context.Context, id int, status string, by int) error {
	f.reports[id].Status, f.reports[id].ResolvedBy = status, &by
	return nil
}
func (f *fakeRepo) UpsertBan(_ context.Context, b domain.Ban) error {
	f.bans[b.UserID] = b
	return nil
}
func (f *fakeRepo) DeleteBan(_ context.Context, _ uuid.UUID, user int) (bool, error) {
	_, ok := f.bans[user]
	delete(f.bans, user)
	return ok, nil
}
func (f *fakeRepo) ListBans(context.Context, uuid.UUID) ([]*domain.Ban, error) { return nil, nil }
func (f *fakeRepo) IsBanned(_ context.Context, _ uuid.UUID, user int) (bool, error) {
	_, ok := f.bans[user]
	return ok, nil
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
	repo.messages, repo.reports, repo.bans = map[uuid.UUID]*domain.MessageRef{}, map[int]*domain.Report{}, map[int]domain.Ban{}
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

func newMessage(repo *fakeRepo, room uuid.UUID, author int) uuid.UUID {
	id := uuid.New()
	a := author
	repo.messages[id] = &domain.MessageRef{RoomID: room, Author: &a}
	return id
}

func TestReportingRules(t *testing.T) {
	_, repo, rt, room := setup()
	uc := NewReportsUseCase(repo, rt)
	ctx := context.Background()
	msg := newMessage(repo, room, otherMemberID)

	if _, err := uc.Report(ctx, msg, strangerID, "offensive"); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Errorf("non-member: err = %v, want ErrForbidden", err)
	}
	if _, err := uc.Report(ctx, msg, otherMemberID, "offensive"); !errors.Is(err, shared_domain.ErrInvalidInput) {
		t.Errorf("own message: err = %v, want ErrInvalidInput", err)
	}
	if _, err := uc.Report(ctx, msg, memberID, "x"); !errors.Is(err, shared_domain.ErrInvalidInput) {
		t.Errorf("too short a reason: err = %v, want ErrInvalidInput", err)
	}
	if _, err := uc.Report(ctx, uuid.New(), memberID, "offensive"); !errors.Is(err, domain.ErrMessageNotFound) {
		t.Errorf("unknown message: err = %v", err)
	}
	if _, err := uc.Report(ctx, msg, memberID, "  offensive  "); err != nil {
		t.Fatalf("valid report: %v", err)
	}
	if _, err := uc.Report(ctx, msg, memberID, "again"); !errors.Is(err, domain.ErrAlreadyReported) {
		t.Errorf("duplicate: err = %v, want ErrAlreadyReported", err)
	}
	if repo.reports[1].Reason != "offensive" {
		t.Error("the reason should be trimmed")
	}
}

func TestOnlyModeratorsSeeAndResolveReports(t *testing.T) {
	_, repo, rt, room := setup()
	uc := NewReportsUseCase(repo, rt)
	ctx := context.Background()
	msg := newMessage(repo, room, otherMemberID)
	if _, err := uc.Report(ctx, msg, memberID, "spam spam"); err != nil {
		t.Fatal(err)
	}

	if _, err := uc.List(ctx, room, memberID, ""); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Errorf("member listing: err = %v, want ErrForbidden", err)
	}
	if list, err := uc.List(ctx, room, modID, ""); err != nil || len(list) != 1 {
		t.Errorf("moderator listing: %v / %d", err, len(list))
	}
	if _, err := uc.List(ctx, room, modID, "weird"); !errors.Is(err, shared_domain.ErrInvalidInput) {
		t.Errorf("bad status: err = %v", err)
	}
	if err := uc.Resolve(ctx, 1, memberID, "resolve", true); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Errorf("member resolving: err = %v, want ErrForbidden", err)
	}
	if err := uc.Resolve(ctx, 1, modID, "explode", false); !errors.Is(err, shared_domain.ErrInvalidInput) {
		t.Errorf("bad action: err = %v", err)
	}
	if err := uc.Resolve(ctx, 99, modID, "resolve", false); !errors.Is(err, domain.ErrReportNotFound) {
		t.Errorf("unknown report: err = %v", err)
	}

	if err := uc.Resolve(ctx, 1, modID, "resolve", true); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if repo.reports[1].Status != domain.ReportResolved || !repo.messages[msg].Deleted {
		t.Error("resolving with delete should close the report and delete the message")
	}
	if len(rt.events) != 1 || rt.events[0] != "delete_message" || len(repo.audit) != 1 {
		t.Errorf("deletion must be broadcast and audited: %v / %v", rt.events, repo.audit)
	}
	if err := uc.Resolve(ctx, 1, modID, "ignore", false); !errors.Is(err, shared_domain.ErrConflict) {
		t.Errorf("already closed: err = %v, want ErrConflict", err)
	}
}

func TestBanRules(t *testing.T) {
	_, repo, rt, room := setup()
	uc := NewBansUseCase(repo, rt)
	ctx := context.Background()

	if err := uc.Ban(ctx, room, memberID, otherMemberID, "", ""); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Errorf("member banning: err = %v", err)
	}
	if err := uc.Ban(ctx, room, modID, adminID, "", ""); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Errorf("moderator banning admin: err = %v", err)
	}
	if err := uc.Ban(ctx, room, globalAdminID, ownerID, "", ""); !errors.Is(err, domain.ErrCannotTargetOwner) {
		t.Errorf("banning the owner: err = %v", err)
	}
	if err := uc.Ban(ctx, room, modID, memberID, "", "soon"); !errors.Is(err, shared_domain.ErrInvalidInput) {
		t.Errorf("bad duration: err = %v", err)
	}
	if err := uc.Ban(ctx, room, modID, modID, "", ""); !errors.Is(err, shared_domain.ErrInvalidInput) {
		t.Errorf("self ban: err = %v", err)
	}

	if err := uc.Ban(ctx, room, modID, memberID, "spam", "24h"); err != nil {
		t.Fatalf("ban: %v", err)
	}
	if banned, _ := repo.IsBanned(ctx, room, memberID); !banned || repo.actors[memberID].IsMember {
		t.Error("banned user should be recorded and removed from the room")
	}
	if repo.bans[memberID].ExpiresAt == nil {
		t.Error("a duration should set an expiry")
	}
	if len(rt.unsubscribed) != 1 || repo.audit[len(repo.audit)-1].Action != domain.ActionMemberBanned {
		t.Error("ban must cut realtime and be audited")
	}

	// banning someone who is not (or no longer) a member works too: it keeps them out
	if err := uc.Ban(ctx, room, adminID, strangerID, "troll", ""); err != nil {
		t.Fatalf("pre-emptive ban: %v", err)
	}
	if repo.bans[strangerID].ExpiresAt != nil {
		t.Error("no duration means a permanent ban")
	}

	if err := uc.Unban(ctx, room, memberID, strangerID); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Errorf("member unbanning: err = %v", err)
	}
	if err := uc.Unban(ctx, room, modID, strangerID); err != nil {
		t.Errorf("moderator unban: %v", err)
	}
	if err := uc.Unban(ctx, room, modID, strangerID); !errors.Is(err, shared_domain.ErrNotFound) {
		t.Errorf("unban twice: err = %v, want ErrNotFound", err)
	}
}
