package usecases

import (
	"context"
	"errors"
	"io"
	"testing"

	auth_domain "backend/internal/auth/domain"
	shared_domain "backend/internal/shared/domain"
	"backend/internal/users/domain"
)

type fakeUsers struct {
	hash    string
	deleted []int
}

func (f *fakeUsers) GetUserByID(context.Context, int) (*auth_domain.User, error) { return nil, nil }
func (f *fakeUsers) DeactivateUser(context.Context, int) error                   { return nil }
func (f *fakeUsers) PasswordHash(context.Context, int) (string, error)           { return f.hash, nil }
func (f *fakeUsers) ListDirectory(context.Context, int, string, int) ([]*domain.DirectoryEntry, error) {
	return nil, nil
}
func (f *fakeUsers) ListDirectoryByIDs(context.Context, int, string, []int, int) ([]*domain.DirectoryEntry, error) {
	return nil, nil
}
func (f *fakeUsers) ExportAccount(context.Context, int, io.Writer) error {
	return nil
}
func (f *fakeUsers) DeleteAccount(_ context.Context, id int) error {
	f.deleted = append(f.deleted, id)
	return nil
}

type spyRealtime struct{ disconnected []int }

func (s *spyRealtime) DisconnectUser(id int) { s.disconnected = append(s.disconnected, id) }

func TestDeleteAccountNeedsTheRightPassword(t *testing.T) {
	passwords := auth_domain.NewPasswordService()
	hash, err := passwords.HashPassword("correct-horse-battery")
	if err != nil {
		t.Fatal(err)
	}
	repo, rt := &fakeUsers{hash: hash}, &spyRealtime{}
	uc := NewDeleteAccountUseCase(repo, passwords, rt)

	if err := uc.Execute(context.Background(), 7, "wrong"); !errors.Is(err, shared_domain.ErrInvalidCredentials) {
		t.Fatalf("wrong password: err = %v, want ErrInvalidCredentials", err)
	}
	if len(repo.deleted) != 0 || len(rt.disconnected) != 0 {
		t.Fatal("nothing may happen with a wrong password")
	}

	if err := uc.Execute(context.Background(), 7, "correct-horse-battery"); err != nil {
		t.Fatalf("right password: %v", err)
	}
	if len(repo.deleted) != 1 || repo.deleted[0] != 7 || len(rt.disconnected) != 1 {
		t.Fatalf("account should be deleted and sockets cut: %v / %v", repo.deleted, rt.disconnected)
	}
}
