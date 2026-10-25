package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"backend/internal/message/domain"
	shared_domain "backend/internal/shared/domain"
	"backend/internal/shared/permissions"

	"github.com/google/uuid"
)

// fakeRepo is an in-memory MessageRepository and ReactionRepository for tests.
type fakeRepo struct {
	members      map[uuid.UUID]map[int]bool
	messages     map[uuid.UUID]*domain.Message
	reactions    int
	reacted      []domain.Reaction
	searched     string
	searchedRoom *uuid.UUID
	roles        map[int]string // overrides the default "member" role
	global       map[int]permissions.GlobalRole
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{members: map[uuid.UUID]map[int]bool{}, messages: map[uuid.UUID]*domain.Message{}}
}

func (f *fakeRepo) addMember(room uuid.UUID, user int) {
	if f.members[room] == nil {
		f.members[room] = map[int]bool{}
	}
	f.members[room][user] = true
}

func (f *fakeRepo) Create(_ context.Context, m *domain.Message) (*domain.Message, error) {
	m.ID = uuid.New()
	f.messages[m.ID] = m
	return m, nil
}
func (f *fakeRepo) Update(_ context.Context, content string, id uuid.UUID) (string, error) {
	f.messages[id].Content = content
	return f.messages[id].RoomID.String(), nil
}
func (f *fakeRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Message, error) {
	if m, ok := f.messages[id]; ok {
		return m, nil
	}
	return nil, domain.ErrMessageNotFound
}
func (f *fakeRepo) SoftDelete(_ context.Context, id uuid.UUID) (string, error) {
	now := time.Now()
	f.messages[id].DeletedAt = &now
	return f.messages[id].RoomID.String(), nil
}
func (f *fakeRepo) ActorIn(_ context.Context, room uuid.UUID, user int) (permissions.Actor, error) {
	a := permissions.Actor{Global: permissions.GlobalUser}
	if g, ok := f.global[user]; ok {
		a.Global = g
	}
	if f.members[room][user] {
		a.IsMember, a.Room = true, permissions.RoomMember
		if role, ok := f.roles[user]; ok {
			a.Room = permissions.RoomRole(role)
		}
	}
	return a, nil
}
func (f *fakeRepo) SetPinned(_ context.Context, id uuid.UUID, pinned bool) error {
	if pinned {
		now := time.Now()
		f.messages[id].PinnedAt = &now
	} else {
		f.messages[id].PinnedAt = nil
	}
	return nil
}
func (f *fakeRepo) FindPinned(_ context.Context, room uuid.UUID) ([]*domain.Message, error) {
	out := []*domain.Message{}
	for _, m := range f.messages {
		if m.RoomID == room && m.PinnedAt != nil {
			out = append(out, m)
		}
	}
	return out, nil
}
func (f *fakeRepo) Search(_ context.Context, user int, q string, room *uuid.UUID, _ int) ([]*domain.Message, error) {
	f.searchedRoom, f.searched = room, q
	return []*domain.Message{}, nil
}
func (f *fakeRepo) FindByRoomID(context.Context, uuid.UUID, int, *uuid.UUID) ([]*domain.Message, error) {
	return []*domain.Message{}, nil
}
func (f *fakeRepo) IsRoomMember(_ context.Context, room uuid.UUID, user int) (bool, error) {
	return f.members[room][user], nil
}
func (f *fakeRepo) Add(_ context.Context, r *domain.Reaction) error {
	f.reactions++
	f.reacted = append(f.reacted, *r)
	return nil
}
func (f *fakeRepo) CountsFor(_ context.Context, msg uuid.UUID, user int, emoji string) (domain.ReactionCounts, error) {
	var c domain.ReactionCounts
	kinds := map[string]bool{}
	for _, r := range f.reacted {
		if r.MessageID != msg {
			continue
		}
		kinds[r.Emoji] = true
		if r.UserID == user {
			c.UserReactions++
		}
		if r.Emoji == emoji {
			c.EmojiInUse = true
			if r.UserID == user {
				c.AlreadyReacted = true
			}
		}
	}
	c.Kinds = len(kinds)
	return c, nil
}
func (f *fakeRepo) FindByMessageID(context.Context, uuid.UUID) ([]*domain.Reaction, error) {
	return nil, nil
}
func (f *fakeRepo) Remove(context.Context, uuid.UUID, int, string) error {
	f.reactions--
	return nil
}

const memberID, outsiderID, adminID = 1, 2, 3

func setup() (*fakeRepo, uuid.UUID, *domain.Message) {
	repo := newFakeRepo()
	room := uuid.New()
	repo.addMember(room, memberID)
	msg, _ := repo.Create(context.Background(), &domain.Message{Content: "hi", AuthorID: memberID, RoomID: room})
	return repo, room, msg
}

func TestCreateMessageRequiresMembership(t *testing.T) {
	repo, room, _ := setup()
	uc := NewCreateMessageUseCase(repo, nil)
	dto := CreateMessageDTO{Content: "hello", RoomID: room.String()}

	if _, err := uc.Execute(context.Background(), dto, outsiderID); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Fatalf("outsider: err = %v, want ErrForbidden", err)
	}
	if _, err := uc.Execute(context.Background(), dto, memberID); err != nil {
		t.Fatalf("member: unexpected err %v", err)
	}
}

func TestListMessagesRequiresMembership(t *testing.T) {
	repo, room, _ := setup()
	uc := NewListMessagesUseCase(repo)

	if _, err := uc.Execute(context.Background(), room.String(), outsiderID, 50, ""); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Fatalf("outsider: err = %v, want ErrForbidden", err)
	}
	if _, err := uc.Execute(context.Background(), room.String(), memberID, 50, ""); err != nil {
		t.Fatalf("member: unexpected err %v", err)
	}
}

func TestUpdateMessageOnlyByAuthor(t *testing.T) {
	repo, room, msg := setup()
	repo.addMember(room, outsiderID) // a member who is not the author
	uc := NewUpdateMessageUseCase(repo, nil)

	if err := uc.Execute(context.Background(), msg.ID, outsiderID, UpdateMessageDTO{Content: "hacked"}); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Fatalf("non-author: err = %v, want ErrForbidden", err)
	}
	if msg.Content != "hi" {
		t.Fatal("content must not change")
	}
	if err := uc.Execute(context.Background(), msg.ID, memberID, UpdateMessageDTO{Content: "edited"}); err != nil {
		t.Fatalf("author: unexpected err %v", err)
	}
	if err := uc.Execute(context.Background(), uuid.New(), memberID, UpdateMessageDTO{Content: "x"}); !errors.Is(err, domain.ErrMessageNotFound) {
		t.Fatalf("missing message: err = %v, want ErrMessageNotFound", err)
	}
}

func TestReactionsRequireMembership(t *testing.T) {
	repo, _, msg := setup()
	add := NewAddReactionUseCase(repo, repo, nil)
	remove := NewRemoveReactionUseCase(repo, repo, nil)

	if err := add.Execute(context.Background(), msg.ID, outsiderID, AddReactionDTO{Emoji: "👍"}); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Fatalf("add outsider: err = %v, want ErrForbidden", err)
	}
	if err := remove.Execute(context.Background(), msg.ID, outsiderID, RemoveReactionDTO{Emoji: "👍"}); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Fatalf("remove outsider: err = %v, want ErrForbidden", err)
	}
	if repo.reactions != 0 {
		t.Fatal("outsider must not change reactions")
	}
	if err := add.Execute(context.Background(), msg.ID, memberID, AddReactionDTO{Emoji: "👍"}); err != nil {
		t.Fatalf("member add: unexpected err %v", err)
	}
	if repo.reactions != 1 {
		t.Fatal("member reaction should be stored")
	}
}

func TestDeleteMessageOnlyByAuthor(t *testing.T) {
	repo, room, msg := setup()
	repo.addMember(room, outsiderID)
	uc := NewDeleteMessageUseCase(repo, nil, nil)

	if err := uc.Execute(context.Background(), msg.ID, outsiderID); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Fatalf("non-author: err = %v, want ErrForbidden", err)
	}
	if msg.DeletedAt != nil {
		t.Fatal("message must not be deleted by a non-author")
	}
	if err := uc.Execute(context.Background(), msg.ID, memberID); err != nil {
		t.Fatalf("author: unexpected err %v", err)
	}
	if err := uc.Execute(context.Background(), msg.ID, memberID); !errors.Is(err, domain.ErrMessageNotFound) {
		t.Fatalf("already deleted: err = %v, want ErrMessageNotFound", err)
	}
}

func TestReplyMustTargetAMessageOfTheSameRoom(t *testing.T) {
	repo, room, msg := setup()
	uc := NewCreateMessageUseCase(repo, nil)
	reply := func(parent string) (*domain.Message, error) {
		return uc.Execute(context.Background(), CreateMessageDTO{Content: "re", RoomID: room.String(), ReplyToMessageID: &parent}, memberID)
	}

	created, err := reply(msg.ID.String())
	if err != nil || created.ReplyToMessageID == nil || *created.ReplyToMessageID != msg.ID {
		t.Fatalf("valid reply: %v / %+v", err, created)
	}

	otherRoom := uuid.New()
	foreign := &domain.Message{ID: uuid.New(), RoomID: otherRoom, AuthorID: memberID, Content: "x"}
	repo.messages[foreign.ID] = foreign
	for name, parent := range map[string]string{"other room": foreign.ID.String(), "unknown": uuid.New().String(), "garbage": "nope"} {
		if _, err := reply(parent); !errors.Is(err, shared_domain.ErrInvalidInput) {
			t.Fatalf("%s: err = %v, want ErrInvalidInput", name, err)
		}
	}

	now := time.Now()
	msg.DeletedAt = &now
	if _, err := reply(msg.ID.String()); !errors.Is(err, shared_domain.ErrInvalidInput) {
		t.Fatalf("deleted parent: err = %v, want ErrInvalidInput", err)
	}
}

func TestPinMessageOnlyByOwnerOrAdmin(t *testing.T) {
	repo, room, msg := setup()
	repo.roles = map[int]string{memberID: "member", adminID: "admin"}
	repo.addMember(room, adminID)
	uc := NewPinMessageUseCase(repo, nil)
	ctx := context.Background()

	for name, user := range map[string]int{"plain member": memberID, "outsider": outsiderID} {
		if err := uc.Execute(ctx, msg.ID, user, true); !errors.Is(err, shared_domain.ErrForbidden) {
			t.Fatalf("%s: err = %v, want ErrForbidden", name, err)
		}
	}
	if msg.PinnedAt != nil {
		t.Fatal("message must not be pinned by unauthorized users")
	}
	if err := uc.Execute(ctx, msg.ID, adminID, true); err != nil || msg.PinnedAt == nil {
		t.Fatalf("admin should pin: err=%v pinned=%v", err, msg.PinnedAt)
	}
	if err := uc.Execute(ctx, msg.ID, adminID, false); err != nil || msg.PinnedAt != nil {
		t.Fatalf("admin should unpin: err=%v pinned=%v", err, msg.PinnedAt)
	}
}

func TestListPinnedRequiresMembership(t *testing.T) {
	repo, room, _ := setup()
	uc := NewListPinnedUseCase(repo)
	if _, err := uc.Execute(context.Background(), room.String(), outsiderID); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Fatalf("outsider: err = %v, want ErrForbidden", err)
	}
}

func TestSearchValidatesInputAndScopesToMembership(t *testing.T) {
	repo, room, _ := setup()
	uc := NewSearchMessagesUseCase(repo)
	ctx := context.Background()

	for name, q := range map[string]string{"empty": "", "one char": "a", "blank": "   "} {
		if _, err := uc.Execute(ctx, memberID, q, "", 0); !errors.Is(err, shared_domain.ErrInvalidInput) {
			t.Fatalf("%s: err = %v, want ErrInvalidInput", name, err)
		}
	}
	if _, err := uc.Execute(ctx, memberID, "hello", "not-a-uuid", 0); !errors.Is(err, shared_domain.ErrInvalidInput) {
		t.Fatalf("bad room id: err = %v", err)
	}
	if _, err := uc.Execute(ctx, outsiderID, "hello", room.String(), 0); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Fatalf("searching a room you are not in: err = %v, want ErrForbidden", err)
	}
	if _, err := uc.Execute(ctx, memberID, "  hello  ", room.String(), 0); err != nil || repo.searched != "hello" || repo.searchedRoom == nil {
		t.Fatalf("member search: err=%v searched=%q room=%v", err, repo.searched, repo.searchedRoom)
	}
}

type spyAuditor struct{ deleted []uuid.UUID }

func (s *spyAuditor) RecordMessageDeletedByModerator(_ context.Context, _ int, _, message uuid.UUID, _ int) error {
	s.deleted = append(s.deleted, message)
	return nil
}

func TestModeratorsCanDeleteOthersMessagesAndItIsAudited(t *testing.T) {
	repo, room, msg := setup()
	repo.addMember(room, adminID)
	repo.roles = map[int]string{adminID: "moderator"}
	auditor := &spyAuditor{}
	uc := NewDeleteMessageUseCase(repo, nil, auditor)

	if err := uc.Execute(context.Background(), msg.ID, outsiderID); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Fatalf("outsider: err = %v, want ErrForbidden", err)
	}
	if err := uc.Execute(context.Background(), msg.ID, adminID); err != nil {
		t.Fatalf("moderator: unexpected err %v", err)
	}
	if msg.DeletedAt == nil || len(auditor.deleted) != 1 {
		t.Fatalf("message should be deleted and audited: %v / %v", msg.DeletedAt, auditor.deleted)
	}
}

func TestAuthorDeletingOwnMessageIsNotAudited(t *testing.T) {
	repo, _, msg := setup()
	auditor := &spyAuditor{}
	if err := NewDeleteMessageUseCase(repo, nil, auditor).Execute(context.Background(), msg.ID, memberID); err != nil {
		t.Fatal(err)
	}
	if len(auditor.deleted) != 0 {
		t.Fatal("own deletions are not moderation actions")
	}
}

func TestGlobalAdminCanPinAnywhere(t *testing.T) {
	repo, _, msg := setup()
	repo.global = map[int]permissions.GlobalRole{outsiderID: permissions.GlobalAdmin}
	if err := NewPinMessageUseCase(repo, nil).Execute(context.Background(), msg.ID, outsiderID, true); err != nil || msg.PinnedAt == nil {
		t.Fatalf("global admin should pin: err=%v", err)
	}
}

func TestUpdateMessageValidatesContentMembershipAndDeleted(t *testing.T) {
	ctx := context.Background()
	repo, room, msg := setup()
	uc := NewUpdateMessageUseCase(repo, nil)

	for name, content := range map[string]string{"blank": "  \n ", "empty": "", "too long": strings.Repeat("a", 4001)} {
		if err := uc.Execute(ctx, msg.ID, memberID, UpdateMessageDTO{Content: content}); !errors.Is(err, shared_domain.ErrInvalidInput) {
			t.Fatalf("%s: err = %v, want ErrInvalidInput", name, err)
		}
	}
	if err := uc.Execute(ctx, msg.ID, memberID, UpdateMessageDTO{Content: strings.Repeat("é", 4000)}); err != nil {
		t.Fatalf("4000 chars must be accepted: %v", err)
	}

	delete(repo.members[room], memberID) // kicked or banned
	if err := uc.Execute(ctx, msg.ID, memberID, UpdateMessageDTO{Content: "still here"}); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Fatalf("former member: err = %v, want ErrForbidden", err)
	}

	repo.addMember(room, memberID)
	now := time.Now()
	msg.DeletedAt = &now
	if err := uc.Execute(ctx, msg.ID, memberID, UpdateMessageDTO{Content: "zombie"}); !errors.Is(err, domain.ErrMessageNotFound) {
		t.Fatalf("deleted message: err = %v, want ErrMessageNotFound", err)
	}
}

func TestCreateMessageRejectsBlankOrTooLongContent(t *testing.T) {
	repo, room, _ := setup()
	uc := NewCreateMessageUseCase(repo, nil)
	for _, content := range []string{"   ", strings.Repeat("a", 4001)} {
		if _, err := uc.Execute(context.Background(), CreateMessageDTO{Content: content, RoomID: room.String()}, memberID); !errors.Is(err, shared_domain.ErrInvalidInput) {
			t.Fatalf("err = %v, want ErrInvalidInput", err)
		}
	}
}

func TestAddReactionValidatesEmojiAndDeletedMessage(t *testing.T) {
	ctx := context.Background()
	repo, _, msg := setup()
	add := NewAddReactionUseCase(repo, repo, nil)

	for name, emoji := range map[string]string{"blank": "  ", "too many runes": "abcdefghi", "too many bytes": strings.Repeat("👍", 9)} {
		if err := add.Execute(ctx, msg.ID, memberID, AddReactionDTO{Emoji: emoji}); !errors.Is(err, shared_domain.ErrInvalidInput) {
			t.Fatalf("%s: err = %v, want ErrInvalidInput", name, err)
		}
	}
	if err := add.Execute(ctx, msg.ID, memberID, AddReactionDTO{Emoji: "👨‍👩‍👧‍👦"}); err != nil {
		t.Fatalf("ZWJ family emoji must be accepted: %v", err)
	}

	now := time.Now()
	msg.DeletedAt = &now
	if err := add.Execute(ctx, msg.ID, memberID, AddReactionDTO{Emoji: "👍"}); !errors.Is(err, domain.ErrMessageNotFound) {
		t.Fatalf("deleted message: err = %v, want ErrMessageNotFound", err)
	}
}

func TestAddReactionCaps(t *testing.T) {
	ctx := context.Background()
	repo, room, msg := setup()
	add := NewAddReactionUseCase(repo, repo, nil)

	for i := 0; i < maxReactionsPerUser; i++ {
		if err := add.Execute(ctx, msg.ID, memberID, AddReactionDTO{Emoji: string(rune(0x1F600 + i))}); err != nil {
			t.Fatalf("reaction %d: %v", i, err)
		}
	}
	if err := add.Execute(ctx, msg.ID, memberID, AddReactionDTO{Emoji: "🦄"}); !errors.Is(err, shared_domain.ErrInvalidInput) {
		t.Fatalf("per-user cap: err = %v, want ErrInvalidInput", err)
	}
	if err := add.Execute(ctx, msg.ID, memberID, AddReactionDTO{Emoji: string(rune(0x1F600))}); err != nil {
		t.Fatalf("re-adding an existing reaction must stay allowed: %v", err)
	}

	// other users fill the remaining distinct kinds, then a new kind is refused
	for u := 10; u < 10+2; u++ {
		repo.addMember(room, u)
		for i := 0; i < maxReactionsPerUser; i++ {
			emoji := string(rune(0x1F300 + (u-10)*maxReactionsPerUser + i))
			if err := add.Execute(ctx, msg.ID, u, AddReactionDTO{Emoji: emoji}); err != nil {
				t.Fatalf("user %d reaction %d: %v", u, i, err)
			}
		}
	}
	// 30 kinds now exist
	repo.addMember(room, 99)
	if err := add.Execute(ctx, msg.ID, 99, AddReactionDTO{Emoji: "🦄"}); !errors.Is(err, shared_domain.ErrInvalidInput) {
		t.Fatalf("kind cap: err = %v, want ErrInvalidInput", err)
	}
	if err := add.Execute(ctx, msg.ID, 99, AddReactionDTO{Emoji: string(rune(0x1F600))}); err != nil {
		t.Fatalf("joining an existing kind must stay allowed: %v", err)
	}
}
