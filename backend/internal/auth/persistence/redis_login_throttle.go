package persistence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"backend/internal/auth/domain"
)

// Redis layout: login:fail:{acct|pair}:<hash> counters (TTL = ThrottleWindow, renewed on every
// failure) and login:wait:{acct|pair}:<hash> keys whose TTL is the remaining delay.

func throttleKeys(identifier, ip string) (acct, pair string) {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(identifier))))
	h := hex.EncodeToString(sum[:16])
	return "acct:" + h, "pair:" + h + ":" + ip
}

func (r *RedisTokenRepository) Check(ctx context.Context, identifier, ip string) (time.Duration, error) {
	acct, pair := throttleKeys(identifier, ip)
	var longest time.Duration
	for _, k := range []string{acct, pair} {
		ttl, err := r.client.PTTL(ctx, "login:wait:"+k).Result()
		if err != nil {
			return 0, err
		}
		if ttl > longest {
			longest = ttl
		}
	}
	return longest, nil
}

func (r *RedisTokenRepository) RecordFailure(ctx context.Context, identifier, ip string) error {
	acct, pair := throttleKeys(identifier, ip)
	for k, threshold := range map[string]int{acct: domain.ThrottleAccountThreshold, pair: domain.ThrottlePairThreshold} {
		n, err := r.client.Incr(ctx, "login:fail:"+k).Result()
		if err != nil {
			return err
		}
		if err := r.client.Expire(ctx, "login:fail:"+k, domain.ThrottleWindow).Err(); err != nil {
			return err
		}
		if d := domain.ThrottleDelay(int(n), threshold); d > 0 {
			if err := r.client.Set(ctx, "login:wait:"+k, 1, d).Err(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *RedisTokenRepository) RecordSuccess(ctx context.Context, identifier, ip string) error {
	_, pair := throttleKeys(identifier, ip)
	// Only the pair is cleared: a success from one address must not erase an attacker's account-wide count.
	return r.client.Del(ctx, "login:fail:"+pair, "login:wait:"+pair).Err()
}
