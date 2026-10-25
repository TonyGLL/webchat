package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/internal/auth/application/dtos"
	"backend/internal/auth/domain"
	shared_domain "backend/internal/shared/domain"
)

type memThrottle struct {
	fails map[string]int
	reset int
}

func (m *memThrottle) Check(_ context.Context, id, ip string) (time.Duration, error) {
	return domain.ThrottleDelay(m.fails[id+"|"+ip], domain.ThrottlePairThreshold), nil
}
func (m *memThrottle) RecordFailure(_ context.Context, id, ip string) error {
	m.fails[id+"|"+ip]++
	return nil
}
func (m *memThrottle) RecordSuccess(_ context.Context, id, ip string) error {
	delete(m.fails, id+"|"+ip)
	m.reset++
	return nil
}

type loginRepo struct {
	domain.AuthRepository
}

func (loginRepo) GetUserByEmailOrUsername(_ context.Context, key string) (*domain.User, error) {
	if key == "ana" {
		return &domain.User{ID: 7, Username: "ana"}, nil
	}
	return nil, shared_domain.ErrNotFound
}
func (loginRepo) ValidateUserPassword(context.Context, int) (string, error) {
	return "hash:secret", nil
}
func (loginRepo) SetLastAccess(context.Context, int) error { return nil }

func TestLoginThrottleSlowsRepeatedFailures(t *testing.T) {
	th := &memThrottle{fails: map[string]int{}}
	uc := NewLoginUseCase(loginRepo{}, newFakeTokens(), plainHasher{}, nil).WithThrottle(th)
	ctx := context.Background()

	for i := 0; i < domain.ThrottlePairThreshold; i++ {
		if _, err := uc.Execute(ctx, dtos.LoginInputDTO{User: "ana", Password: "wrong", IP: "1.1.1.1"}); !errors.Is(err, shared_domain.ErrInvalidCredentials) {
			t.Fatalf("attempt %d: %v", i, err)
		}
	}
	var throttled *domain.ThrottledError
	// the correct password is not even checked while cooling down
	if _, err := uc.Execute(ctx, dtos.LoginInputDTO{User: "ana", Password: "secret", IP: "1.1.1.1"}); !errors.As(err, &throttled) || throttled.RetryAfter <= 0 {
		t.Fatalf("expected a ThrottledError, got %v", err)
	}
	// another address is not affected by that pair's delay
	if _, err := uc.Execute(ctx, dtos.LoginInputDTO{User: "ana", Password: "wrong", IP: "2.2.2.2"}); !errors.Is(err, shared_domain.ErrInvalidCredentials) {
		t.Fatalf("other IP: %v", err)
	}
}

func TestLoginThrottleCountsUnknownAccounts(t *testing.T) {
	th := &memThrottle{fails: map[string]int{}}
	uc := NewLoginUseCase(loginRepo{}, newFakeTokens(), plainHasher{}, nil).WithThrottle(th)
	_, _ = uc.Execute(context.Background(), dtos.LoginInputDTO{User: "ghost", Password: "x", IP: "1.1.1.1"})
	if th.fails["ghost|1.1.1.1"] != 1 {
		t.Fatal("unknown accounts must count too, so the throttle does not reveal which exist")
	}
}

func TestThrottleDelayIsProgressiveAndCapped(t *testing.T) {
	cases := []struct {
		failures, threshold int
		want                time.Duration
	}{
		{0, 5, 0}, {4, 5, 0}, {5, 5, time.Second}, {6, 5, 2 * time.Second}, {7, 5, 4 * time.Second},
		{15, 5, 5 * time.Minute}, {1000, 5, 5 * time.Minute}, {20, 20, time.Second},
	}
	for _, c := range cases {
		if got := domain.ThrottleDelay(c.failures, c.threshold); got != c.want {
			t.Errorf("ThrottleDelay(%d,%d) = %v, want %v", c.failures, c.threshold, got, c.want)
		}
	}
}
