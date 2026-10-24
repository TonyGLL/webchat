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
func (f *fakeRooms) FindOrCreateDirect(context.Context, int, int) (*domain.Room, error) {
	return &domain.Room{ID: uuid.New(), IsDirect: true}, nil
}
func (f *fakeRooms) FindRoomsByUserID(context.Context, int) ([]*domain.Room, error) { return nil, nil }
func (f *fakeRooms) ListDiscoverable(context.Context, int) ([]*domain.RoomSummary, error) {
	return nil, nil
}

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
	uc := NewJoinRoomUseCase(rooms, members, nil, nil)
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
	accept := NewAcceptInviteUseCase(invites, members, nil, nil)
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

func (f *fakeMembers) RemoveMember(_ context.Context, room uuid.UUID, user int) error {
	delete(f.members[room], user)
	return nil
}

type spySubscriber struct{ unsubscribed []int }

func (s *spySubscriber) SubscribeUser(int, string) {}
func (s *spySubscriber) UnsubscribeUser(user int, _ string) {
	s.unsubscribed = append(s.unsubscribed, user)
}

func TestLeaveRoom(t *testing.T) {
	room := uuid.New()
	members := &fakeMembers{members: map[uuid.UUID]map[int]domain.RoomRole{room: {1: domain.RoleOwner, 2: domain.RoleMember}}}
	spy := &spySubscriber{}
	uc := NewLeaveRoomUseCase(members, spy)
	ctx := context.Background()

	if err := uc.Execute(ctx, room, 1); !errors.Is(err, ErrOwnerCannotLeave) {
		t.Fatalf("owner: err = %v, want ErrOwnerCannotLeave", err)
	}
	if err := uc.Execute(ctx, room, 3); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Fatalf("outsider: err = %v, want ErrForbidden", err)
	}
	if err := uc.Execute(ctx, room, 2); err != nil {
		t.Fatalf("member: unexpected err %v", err)
	}
	if _, still := members.members[room][2]; still {
		t.Fatal("member should be removed")
	}
	if len(spy.unsubscribed) != 1 || spy.unsubscribed[0] != 2 {
		t.Fatalf("websocket unsubscribe not requested: %v", spy.unsubscribed)
	}
}

func (f *fakeMembers) ListMembers(_ context.Context, room uuid.UUID) ([]*domain.MemberInfo, error) {
	out := []*domain.MemberInfo{}
	for id, role := range f.members[room] {
		out = append(out, &domain.MemberInfo{UserID: id, Role: role})
	}
	return out, nil
}

type fakePresence struct{ online map[int]bool }

func (p fakePresence) OnlineUsers(context.Context, []int) (map[int]bool, error) { return p.online, nil }

func TestListMembersShowsPresenceToMembersOnly(t *testing.T) {
	room := uuid.New()
	members := &fakeMembers{members: map[uuid.UUID]map[int]domain.RoomRole{room: {1: domain.RoleOwner, 2: domain.RoleMember}}}
	uc := NewListMembersUseCase(members, fakePresence{online: map[int]bool{2: true}})

	if _, err := uc.Execute(context.Background(), room, 3); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Fatalf("outsider: err = %v, want ErrForbidden", err)
	}
	list, err := uc.Execute(context.Background(), room, 1)
	if err != nil || len(list) != 2 {
		t.Fatalf("member: err=%v len=%d", err, len(list))
	}
	for _, m := range list {
		if m.Online != (m.UserID == 2) {
			t.Fatalf("user %d online=%v", m.UserID, m.Online)
		}
	}
}

type spyDirectSubscriber struct{ subscribed []int }

func (s *spyDirectSubscriber) SubscribeUser(user int, _ string) {
	s.subscribed = append(s.subscribed, user)
}
func (s *spyDirectSubscriber) UnsubscribeUser(int, string) {}

func TestOpenDirect(t *testing.T) {
	rooms := &fakeRooms{rooms: map[uuid.UUID]*domain.Room{}}
	members := &fakeMembers{members: map[uuid.UUID]map[int]domain.RoomRole{}}
	spy := &spyDirectSubscriber{}
	uc := NewOpenDirectUseCase(rooms, members, spy)
	ctx := context.Background()

	if _, err := uc.Execute(ctx, 1, 1); !errors.Is(err, shared_domain.ErrInvalidInput) {
		t.Fatalf("self: err = %v, want ErrInvalidInput", err)
	}

	room, err := uc.Execute(ctx, 1, 2)
	if err != nil || !room.IsDirect {
		t.Fatalf("open: room=%+v err=%v", room, err)
	}
	if _, ok := members.members[room.ID][1]; !ok {
		t.Fatal("caller must be (re)added as a member")
	}
	if len(spy.subscribed) != 2 {
		t.Fatalf("both users must be subscribed to live events, got %v", spy.subscribed)
	}
}

type fakeBans struct{ banned map[int]bool }

func (b fakeBans) IsBanned(_ context.Context, _ uuid.UUID, user int) (bool, error) {
	return b.banned[user], nil
}

func TestBannedUsersCannotJoin(t *testing.T) {
	pub := uuid.New()
	rooms := &fakeRooms{rooms: map[uuid.UUID]*domain.Room{pub: {ID: pub}}}
	members := &fakeMembers{members: map[uuid.UUID]map[int]domain.RoomRole{}}
	uc := NewJoinRoomUseCase(rooms, members, nil, fakeBans{banned: map[int]bool{9: true}})

	if err := uc.Execute(context.Background(), pub, 9); !errors.Is(err, ErrBannedFromRoom) {
		t.Fatalf("banned user: err = %v, want ErrBannedFromRoom", err)
	}
	if _, joined := members.members[pub][9]; joined {
		t.Fatal("a banned user must not become a member")
	}
	if err := uc.Execute(context.Background(), pub, 1); err != nil {
		t.Fatalf("normal user: %v", err)
	}
}
