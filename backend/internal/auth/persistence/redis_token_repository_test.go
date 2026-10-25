package persistence

import (
	"context"
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	"backend/internal/auth/domain"
	shared_domain "backend/internal/shared/domain"

	"github.com/redis/go-redis/v9"
)

// Needs a real Redis: set TEST_REDIS_ADDR (a throwaway instance) to run it.
func newTestRepo(t *testing.T) (*RedisTokenRepository, context.Context) {
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("TEST_REDIS_ADDR not set")
	}
	client := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { client.Close() })
	return &RedisTokenRepository{client: client}, context.Background()
}

func uniq(prefix string) string { return prefix + strconv.FormatInt(time.Now().UnixNano(), 36) }

func TestRedisRotationReplayAndReuse(t *testing.T) {
	r, ctx := newTestRepo(t)
	first, fam := uniq("first-"), uniq("fam-")
	if err := r.StoreToken(ctx, 7, first, "refresh_token", time.Minute); err != nil {
		t.Fatal(err)
	}
	rotate := func(presented, newID string, legacy bool) (domain.RotateOutcome, string) {
		o, succ, err := r.RotateRefreshToken(ctx, domain.RotateRequest{UserID: 7, Family: fam, PresentedID: presented, NewID: newID,
			Legacy: legacy, Grace: 2 * time.Second, TTL: time.Minute})
		if err != nil {
			t.Fatal(err)
		}
		return o, succ
	}
	second := uniq("second-")
	if o, _ := rotate(first, second, true); o != domain.RotateOK {
		t.Fatalf("legacy first rotation: %v", o)
	}
	if o, succ := rotate(first, uniq("other-"), true); o != domain.RotateReplay || succ != second {
		t.Fatalf("replay must return the same successor, got %v %q", o, succ)
	}
	if o, _ := rotate(second, uniq("third-"), false); o != domain.RotateOK {
		t.Fatalf("family rotation: %v", o)
	}
	time.Sleep(2500 * time.Millisecond) // grace over
	if o, _ := rotate(first, uniq("x-"), false); o != domain.RotateReuse {
		t.Fatalf("reuse after the grace must be detected, got %v", o)
	}
	if o, _ := rotate(second, uniq("y-"), false); o != domain.RotateInvalid {
		t.Fatalf("the family must be revoked, got %v", o)
	}
}

func TestRedisConsumeIsSingleUseAndExclusiveTokensReplaceEachOther(t *testing.T) {
	r, ctx := newTestRepo(t)
	a, b := uniq("a-"), uniq("b-")
	if err := r.StoreExclusiveToken(ctx, 9, a, "password_reset_test", time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := r.StoreExclusiveToken(ctx, 9, b, "password_reset_test", time.Minute); err != nil {
		t.Fatal(err)
	}
	if _, err := r.ConsumeToken(ctx, "password_reset_test", a); !errors.Is(err, shared_domain.ErrNotFound) {
		t.Fatalf("the replaced token must be gone, got %v", err)
	}
	if id, err := r.ConsumeToken(ctx, "password_reset_test", b); err != nil || id != 9 {
		t.Fatalf("consume: %v %d", err, id)
	}
	if _, err := r.ConsumeToken(ctx, "password_reset_test", b); !errors.Is(err, shared_domain.ErrNotFound) {
		t.Fatalf("second consume must fail, got %v", err)
	}
}
