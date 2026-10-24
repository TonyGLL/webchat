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
