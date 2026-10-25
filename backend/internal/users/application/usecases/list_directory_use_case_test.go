package usecases

import (
	"context"
	"errors"
	"io"
	"sort"
	"testing"

	auth_domain "backend/internal/auth/domain"
	"backend/internal/users/domain"
)

// dirRepo is an in-memory directory sorted by id, where id order equals name order.
type dirRepo struct{ total int }

func (r *dirRepo) GetUserByID(context.Context, int) (*auth_domain.User, error) { return nil, nil }
func (r *dirRepo) DeactivateUser(context.Context, int) error                   { return nil }
func (r *dirRepo) PasswordHash(context.Context, int) (string, error)           { return "", nil }
func (r *dirRepo) ExportAccount(context.Context, int, io.Writer) error {
	return nil
}
func (r *dirRepo) DeleteAccount(context.Context, int) error { return nil }

func (r *dirRepo) ListDirectory(_ context.Context, selfID int, _ string, limit int) ([]*domain.DirectoryEntry, error) {
	out := []*domain.DirectoryEntry{}
	for id := 1; id <= r.total && len(out) < limit; id++ {
		if id != selfID {
			out = append(out, &domain.DirectoryEntry{ID: id})
		}
	}
	return out, nil
}

func (r *dirRepo) ListDirectoryByIDs(_ context.Context, selfID int, _ string, ids []int, limit int) ([]*domain.DirectoryEntry, error) {
	sorted := append([]int(nil), ids...)
	sort.Ints(sorted)
	out := []*domain.DirectoryEntry{}
	for _, id := range sorted {
		if id != selfID && id >= 1 && id <= r.total && len(out) < limit {
			out = append(out, &domain.DirectoryEntry{ID: id})
		}
	}
	return out, nil
}

type fakePresence struct {
	online []int
	err    error
}

func (p *fakePresence) OnlineUserIDs(_ context.Context, max int) ([]int, error) {
	if p.err != nil {
		return nil, p.err
	}
	if len(p.online) > max {
		return p.online[:max], nil
	}
	return p.online, nil
}

func (p *fakePresence) OnlineUsers(_ context.Context, ids []int) (map[int]bool, error) {
	m := map[int]bool{}
	for _, id := range ids {
		for _, o := range p.online {
			if id == o {
				m[id] = true
			}
		}
	}
	return m, nil
}

func ids(entries []*domain.DirectoryEntry) []int {
	out := make([]int, len(entries))
	for i, e := range entries {
		out[i] = e.ID
	}
	return out
}

func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestDirectoryOnlineBeyondFirstPageComeFirst(t *testing.T) {
	uc := NewListDirectoryUseCase(&dirRepo{total: 500}, &fakePresence{online: []int{400, 250}})
	got, err := uc.Execute(context.Background(), 1, "", 5)
	if err != nil {
		t.Fatal(err)
	}
	if want := []int{250, 400, 2, 3, 4}; !equal(ids(got), want) {
		t.Fatalf("ids = %v, want %v", ids(got), want)
	}
	if !got[0].Online || !got[1].Online || got[2].Online {
		t.Fatalf("online flags wrong: %+v %+v %+v", got[0], got[1], got[2])
	}
}

func TestDirectoryOnlineAlreadyInFirstPageNotDuplicated(t *testing.T) {
	uc := NewListDirectoryUseCase(&dirRepo{total: 500}, &fakePresence{online: []int{3}})
	got, _ := uc.Execute(context.Background(), 1, "", 4)
	if want := []int{3, 2, 4, 5}; !equal(ids(got), want) {
		t.Fatalf("ids = %v, want %v", ids(got), want)
	}
}

func TestDirectoryOnlineFillsWholeLimit(t *testing.T) {
	uc := NewListDirectoryUseCase(&dirRepo{total: 500}, &fakePresence{online: []int{10, 20, 30, 40}})
	got, _ := uc.Execute(context.Background(), 1, "", 3)
	if want := []int{10, 20, 30}; !equal(ids(got), want) {
		t.Fatalf("ids = %v, want %v", ids(got), want)
	}
}

func TestDirectorySelfIsNeverListed(t *testing.T) {
	uc := NewListDirectoryUseCase(&dirRepo{total: 10}, &fakePresence{online: []int{1, 7}})
	got, _ := uc.Execute(context.Background(), 1, "", 3)
	if want := []int{7, 2, 3}; !equal(ids(got), want) {
		t.Fatalf("ids = %v, want %v", ids(got), want)
	}
}

func TestDirectoryWithoutPresenceStaysAlphabetical(t *testing.T) {
	for name, p := range map[string]PresenceChecker{
		"nil":   nil,
		"error": &fakePresence{err: errors.New("redis down")},
	} {
		uc := NewListDirectoryUseCase(&dirRepo{total: 10}, p)
		got, err := uc.Execute(context.Background(), 1, "", 3)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if want := []int{2, 3, 4}; !equal(ids(got), want) {
			t.Fatalf("%s: ids = %v, want %v", name, ids(got), want)
		}
	}
}
