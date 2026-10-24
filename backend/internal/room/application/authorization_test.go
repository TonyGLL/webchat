package application

import (
	"context"
	"errors"
	"testing"

	"backend/internal/room/domain"
	shared_domain "backend/internal/shared/domain"

	"github.com/google/uuid"
)

type fakeRooms struct{ rooms map[uuid.UUID]*domain.Room }

func (f *fakeRooms) Create(_ context.Context, r *domain.Room) (*domain.Room, error) { return r, nil }
func (f *fakeRooms) FindByID(_ context.Context, id uuid.UUID) (*domain.Room, error) {
	return f.rooms[id], nil
}
func (f *fakeRooms) FindRoomsByUserID(context.Context, int) ([]*domain.Room, error) { return nil, nil }

type fakeMembers struct {
	members map[uuid.UUID]map[int]domain.RoomRole
}

func (f *fakeMembers) AddMember(_ context.Context, m *domain.RoomMember) error {
	if f.members[m.RoomID] == nil {
		f.members[m.RoomID] = map[int]domain.RoomRole{}
	}
	f.members[m.RoomID][m.UserID] = m.Role
	return nil
}
func (f *fakeMembers) FindMember(_ context.Context, room uuid.UUID, user int) (*domain.RoomMember, error) {
	role, ok := f.members[room][user]
	if !ok {
		return nil, nil
	}
	return &domain.RoomMember{RoomID: room, UserID: user, Role: role}, nil
}

type fakeInvites struct{ invites map[string]*domain.Invite }

func (f *fakeInvites) Create(_ context.Context, i *domain.Invite) (*domain.Invite, error) {
	f.invites[i.Code] = i
	return i, nil
}
func (f *fakeInvites) GetByCode(_ context.Context, code string) (*domain.Invite, error) {
	if i, ok := f.invites[code]; ok {
		return i, nil
	}
	return nil, domain.ErrInviteNotFound
}

func TestJoinRoom(t *testing.T) {
	pub, prv := uuid.New(), uuid.New()
	rooms := &fakeRooms{rooms: map[uuid.UUID]*domain.Room{
		pub: {ID: pub},
		prv: {ID: prv, IsPrivate: true},
	}}
	members := &fakeMembers{members: map[uuid.UUID]map[int]domain.RoomRole{}}
	uc := NewJoinRoomUseCase(rooms, members, nil)
	ctx := context.Background()

	if err := uc.Execute(ctx, pub, 1); err != nil {
		t.Fatalf("public room: unexpected err %v", err)
	}
	if err := uc.Execute(ctx, pub, 1); !errors.Is(err, ErrAlreadyMember) {
		t.Fatalf("second join: err = %v, want ErrAlreadyMember", err)
	}
	if err := uc.Execute(ctx, prv, 1); !errors.Is(err, ErrRoomIsPrivate) {
		t.Fatalf("private room: err = %v, want ErrRoomIsPrivate", err)
	}
	if err := uc.Execute(ctx, uuid.New(), 1); !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("missing room: err = %v, want ErrRoomNotFound", err)
	}
}

func TestInviteFlow(t *testing.T) {
	room := uuid.New()
	members := &fakeMembers{members: map[uuid.UUID]map[int]domain.RoomRole{
		room: {1: domain.RoleOwner, 2: domain.RoleAdmin, 3: domain.RoleMember},
	}}
	invites := &fakeInvites{invites: map[string]*domain.Invite{}}
	create := NewCreateInviteUseCase(invites, &fakeRooms{}, members)
	accept := NewAcceptInviteUseCase(invites, members, nil)
	ctx := context.Background()
	dto := CreateInviteDTO{RoomID: room.String(), ExpiresIn: "24h"}

	for _, user := range []int{99, 3} { // outsider and plain member
		if _, err := create.Execute(ctx, dto, user); !errors.Is(err, shared_domain.ErrForbidden) {
			t.Fatalf("user %d: err = %v, want ErrForbidden", user, err)
		}
	}
	for _, user := range []int{1, 2} { // owner and admin
		if _, err := create.Execute(ctx, dto, user); err != nil {
			t.Fatalf("user %d: unexpected err %v", user, err)
		}
	}

	var code string
	for c := range invites.invites {
		code = c
		break
	}
	if _, err := accept.Execute(ctx, AcceptInviteDTO{Code: code}, 50); err != nil {
		t.Fatalf("accept: unexpected err %v", err)
	}
	if _, err := accept.Execute(ctx, AcceptInviteDTO{Code: code}, 50); !errors.Is(err, ErrAlreadyMember) {
		t.Fatalf("accept twice: err = %v, want ErrAlreadyMember", err)
	}
	if _, err := accept.Execute(ctx, AcceptInviteDTO{Code: "nope"}, 51); !errors.Is(err, domain.ErrInviteNotFound) {
		t.Fatalf("bad code: err = %v, want ErrInviteNotFound", err)
	}
}

func (f *fakeMembers) MarkRead(_ context.Context, room uuid.UUID, user int) (bool, error) {
	_, ok := f.members[room][user]
	return ok, nil
}

func TestMarkReadRequiresMembership(t *testing.T) {
	room := uuid.New()
	members := &fakeMembers{members: map[uuid.UUID]map[int]domain.RoomRole{room: {1: domain.RoleMember}}}
	uc := NewMarkRoomReadUseCase(members)

	if err := uc.Execute(context.Background(), room, 1); err != nil {
		t.Fatalf("member should be able to mark read: %v", err)
	}
	if err := uc.Execute(context.Background(), room, 2); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Fatalf("outsider should be forbidden, got %v", err)
	}
}
