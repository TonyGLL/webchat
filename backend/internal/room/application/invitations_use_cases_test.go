package application

import (
	"context"
	"errors"
	"testing"

	"backend/internal/room/domain"
	shared_domain "backend/internal/shared/domain"

	"github.com/google/uuid"
)

type fakeInvitations struct {
	pending map[uuid.UUID]*domain.Invitation
}

func (f *fakeInvitations) Create(_ context.Context, roomID uuid.UUID, inviterID, inviteeID int) (*domain.Invitation, error) {
	inv := &domain.Invitation{ID: uuid.New(), RoomID: roomID, InviterID: inviterID, InviteeID: inviteeID}
	f.pending[inv.ID] = inv
	return inv, nil
}
func (f *fakeInvitations) ListByRoom(_ context.Context, roomID uuid.UUID) ([]*domain.Invitation, error) {
	var out []*domain.Invitation
	for _, i := range f.pending {
		if i.RoomID == roomID {
			out = append(out, i)
		}
	}
	return out, nil
}
func (f *fakeInvitations) Cancel(_ context.Context, roomID, id uuid.UUID) (bool, error) {
	if i := f.pending[id]; i != nil && i.RoomID == roomID {
		delete(f.pending, id)
		return true, nil
	}
	return false, nil
}
func (f *fakeInvitations) ListPending(context.Context, int) ([]*domain.Invitation, error) {
	return nil, nil
}
func (f *fakeInvitations) FindPending(_ context.Context, id uuid.UUID) (*domain.Invitation, error) {
	return f.pending[id], nil
}
func (f *fakeInvitations) Resolve(_ context.Context, id uuid.UUID, _ bool) (bool, error) {
	delete(f.pending, id)
	return true, nil
}

func newInvitationFixture(role domain.RoomRole) (roomID uuid.UUID, members *fakeMembers, invitations *fakeInvitations, rooms *fakeRooms) {
	roomID = uuid.New()
	members = &fakeMembers{members: map[uuid.UUID]map[int]domain.RoomRole{roomID: {1: role}}}
	invitations = &fakeInvitations{pending: map[uuid.UUID]*domain.Invitation{}}
	rooms = &fakeRooms{rooms: map[uuid.UUID]*domain.Room{roomID: {ID: roomID}}}
	return
}

func TestInviteUserRequiresInvitePower(t *testing.T) {
	for role, allowed := range map[domain.RoomRole]bool{domain.RoleOwner: true, domain.RoleAdmin: true, domain.RoleModerator: false, domain.RoleMember: false} {
		roomID, members, invitations, rooms := newInvitationFixture(role)
		uc := NewInviteUserUseCase(invitations, rooms, members, nil)
		_, err := uc.Execute(context.Background(), roomID, 1, 2)
		if allowed && err != nil {
			t.Errorf("%s should be able to invite, got %v", role, err)
		}
		if !allowed && !errors.Is(err, shared_domain.ErrForbidden) {
			t.Errorf("%s should not be able to invite, got %v", role, err)
		}
	}
}

func TestInviteUserRejectsSelfAndExistingMembers(t *testing.T) {
	roomID, members, invitations, rooms := newInvitationFixture(domain.RoleOwner)
	members.members[roomID][2] = domain.RoleMember
	uc := NewInviteUserUseCase(invitations, rooms, members, nil)
	if _, err := uc.Execute(context.Background(), roomID, 1, 1); !errors.Is(err, shared_domain.ErrInvalidInput) {
		t.Errorf("inviting yourself: got %v", err)
	}
	if _, err := uc.Execute(context.Background(), roomID, 1, 2); !errors.Is(err, ErrAlreadyMember) {
		t.Errorf("inviting a member: got %v", err)
	}
}

func TestRespondInvitation(t *testing.T) {
	roomID, members, invitations, _ := newInvitationFixture(domain.RoleOwner)
	inv, _ := invitations.Create(context.Background(), roomID, 1, 2)
	uc := NewRespondInvitationUseCase(invitations, members, nil, nil)

	if err := uc.Execute(context.Background(), inv.ID, 3, true); !errors.Is(err, shared_domain.ErrNotFound) {
		t.Errorf("someone else's invitation: got %v", err)
	}
	if err := uc.Execute(context.Background(), inv.ID, 2, true); err != nil {
		t.Fatalf("accept: %v", err)
	}
	if members.members[roomID][2] != domain.RoleMember {
		t.Error("accepting should add the user as a member")
	}
	if err := uc.Execute(context.Background(), inv.ID, 2, true); !errors.Is(err, shared_domain.ErrNotFound) {
		t.Errorf("answering twice: got %v", err)
	}
}

func TestDeclineDoesNotAddMember(t *testing.T) {
	roomID, members, invitations, _ := newInvitationFixture(domain.RoleOwner)
	inv, _ := invitations.Create(context.Background(), roomID, 1, 2)
	if err := NewRespondInvitationUseCase(invitations, members, nil, nil).Execute(context.Background(), inv.ID, 2, false); err != nil {
		t.Fatal(err)
	}
	if _, ok := members.members[roomID][2]; ok {
		t.Error("declining must not add the user")
	}
}

func TestAcceptNeedsInviterToStillHaveInvitePower(t *testing.T) {
	for name, setup := range map[string]func(members map[int]domain.RoomRole){
		"demoted": func(m map[int]domain.RoomRole) { m[1] = domain.RoleMember },
		"left":    func(m map[int]domain.RoomRole) { delete(m, 1) },
	} {
		roomID, members, invitations, _ := newInvitationFixture(domain.RoleAdmin)
		inv, _ := invitations.Create(context.Background(), roomID, 1, 2)
		setup(members.members[roomID])
		err := NewRespondInvitationUseCase(invitations, members, nil, nil).Execute(context.Background(), inv.ID, 2, true)
		if !errors.Is(err, shared_domain.ErrNotFound) {
			t.Errorf("%s inviter: err = %v, want ErrNotFound", name, err)
		}
		if _, ok := members.members[roomID][2]; ok {
			t.Errorf("%s inviter: the invitee must not join", name)
		}
	}
}

func TestListAndCancelRoomInvitationsNeedInvitePower(t *testing.T) {
	roomID, members, invitations, _ := newInvitationFixture(domain.RoleAdmin)
	members.members[roomID][5] = domain.RoleModerator
	inv, _ := invitations.Create(context.Background(), roomID, 1, 2)
	ctx := context.Background()
	list := NewListRoomInvitationsUseCase(invitations, members)
	cancel := NewCancelInvitationUseCase(invitations, members)

	if _, err := list.Execute(ctx, roomID, 5); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Errorf("moderator listing: %v", err)
	}
	if err := cancel.Execute(ctx, roomID, inv.ID, 5); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Errorf("moderator cancelling: %v", err)
	}
	if got, err := list.Execute(ctx, roomID, 1); err != nil || len(got) != 1 {
		t.Errorf("admin listing: %v %v", got, err)
	}
	if err := cancel.Execute(ctx, roomID, inv.ID, 1); err != nil {
		t.Errorf("admin cancelling: %v", err)
	}
	if err := cancel.Execute(ctx, roomID, inv.ID, 1); !errors.Is(err, shared_domain.ErrNotFound) {
		t.Errorf("cancelling twice: %v", err)
	}
}
