package application

import (
	"context"
	"errors"
	"testing"

	"backend/internal/message/domain"
	shared_domain "backend/internal/shared/domain"

	"github.com/google/uuid"
)

// fakeRepo is an in-memory MessageRepository and ReactionRepository for tests.
type fakeRepo struct {
	members   map[uuid.UUID]map[int]bool
	messages  map[uuid.UUID]*domain.Message
	reactions int
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
func (f *fakeRepo) FindByRoomID(context.Context, uuid.UUID, int, int) ([]*domain.Message, error) {
	return []*domain.Message{}, nil
}
func (f *fakeRepo) IsRoomMember(_ context.Context, room uuid.UUID, user int) (bool, error) {
	return f.members[room][user], nil
}
func (f *fakeRepo) Add(context.Context, *domain.Reaction) error { f.reactions++; return nil }
func (f *fakeRepo) FindByMessageID(context.Context, uuid.UUID) ([]*domain.Reaction, error) {
	return nil, nil
}
func (f *fakeRepo) Remove(context.Context, uuid.UUID, int, string) error {
	f.reactions--
	return nil
}

const memberID, outsiderID = 1, 2

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

	if _, err := uc.Execute(context.Background(), room.String(), outsiderID, 1, 50); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Fatalf("outsider: err = %v, want ErrForbidden", err)
	}
	if _, err := uc.Execute(context.Background(), room.String(), memberID, 1, 50); err != nil {
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
