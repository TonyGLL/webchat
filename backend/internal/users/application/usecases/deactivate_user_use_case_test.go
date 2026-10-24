package usecases

import (
	"context"
	"testing"
	"time"

	auth_domain "backend/internal/auth/domain"
	shared_domain "backend/internal/shared/domain"
)

type spyRevoker struct{ revoked []int }

func (s *spyRevoker) RevokeUserSessions(_ context.Context, id int, _ time.Duration) error {
	s.revoked = append(s.revoked, id)
	return nil
}

type deactivatingUsers struct {
	fakeUsers
	deactivated []int
}

func (f *deactivatingUsers) DeactivateUser(_ context.Context, id int) error {
	f.deactivated = append(f.deactivated, id)
	return nil
}

func TestDeactivateNeedsPasswordAndEndsSessions(t *testing.T) {
	passwords := auth_domain.NewPasswordService()
	hash, err := passwords.HashPassword("correct-horse-battery")
	if err != nil {
		t.Fatal(err)
	}
	users := &deactivatingUsers{fakeUsers: fakeUsers{hash: hash}}
	revoker, rt := &spyRevoker{}, &spyRealtime{}
	uc := NewDeactivateUserUseCase(users, passwords, revoker, rt)

	if err := uc.Execute(context.Background(), 5, "wrong"); err != shared_domain.ErrInvalidCredentials {
		t.Fatalf("wrong password must be refused, got %v", err)
	}
	if len(users.deactivated) != 0 || len(revoker.revoked) != 0 || len(rt.disconnected) != 0 {
		t.Fatal("nothing may happen with a wrong password")
	}

	if err := uc.Execute(context.Background(), 5, "correct-horse-battery"); err != nil {
		t.Fatal(err)
	}
	if len(users.deactivated) != 1 || len(revoker.revoked) != 1 || len(rt.disconnected) != 1 {
		t.Fatalf("deactivated=%v revoked=%v disconnected=%v", users.deactivated, revoker.revoked, rt.disconnected)
	}
}
