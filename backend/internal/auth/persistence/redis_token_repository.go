package persistence

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"backend/internal/auth/domain"
	shared_domain "backend/internal/shared/domain"

	"github.com/redis/go-redis/v9"
)

// RedisTokenRepository is a Redis-based implementation of the TokenRepository.
type RedisTokenRepository struct {
	client *redis.Client
}

// NewRedisTokenRepository creates a new instance of RedisTokenRepository.
func NewRedisTokenRepository(client *redis.Client) domain.TokenRepository {
	return &RedisTokenRepository{
		client: client,
	}
}

func (r *RedisTokenRepository) StoreToken(ctx context.Context, userID int, tokenID string, keyIn string, expiresIn time.Duration) error {
	key := r.getRedisKey(keyIn, tokenID)
	err := r.client.Set(ctx, key, userID, expiresIn).Err()
	if err != nil {
		return fmt.Errorf("could not store refresh token in redis: %w", err)
	}
	return nil
}

// GetUserIDByRefreshToken retrieves the user ID associated with the given refresh token ID.
// If the token is not found in Redis, it returns a domain-specific ErrNotFound.
func (r *RedisTokenRepository) GetToken(ctx context.Context, keyIn string, tokenID string) (int, error) {
	key := r.getRedisKey(keyIn, tokenID)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, shared_domain.ErrNotFound
		}
		return 0, fmt.Errorf("could not get refresh token from redis: %w", err)
	}

	userID, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("could not parse userID from redis value: %w", err)
	}

	return userID, nil
}

// DeleteRefreshToken removes a refresh token from Redis.
func (r *RedisTokenRepository) DeleteToken(ctx context.Context, keyIn string, tokenID string) error {
	key := r.getRedisKey(keyIn, tokenID)
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("could not delete refresh token from redis: %w", err)
	}
	return nil
}

// ConsumeToken reads and deletes a token in one step (GETDEL).
func (r *RedisTokenRepository) ConsumeToken(ctx context.Context, keyIn string, tokenID string) (int, error) {
	val, err := r.client.GetDel(ctx, r.getRedisKey(keyIn, tokenID)).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, shared_domain.ErrNotFound
		}
		return 0, fmt.Errorf("could not consume token from redis: %w", err)
	}
	userID, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("could not parse userID from redis value: %w", err)
	}
	return userID, nil
}

// storeExclusiveScript drops the user's previous token for the key and stores the new one.
// KEYS: 1 new token, 2 per-user pointer. ARGV: 1 userID, 2 tokenID, 3 ttl seconds, 4 key prefix.
var storeExclusiveScript = redis.NewScript(`
local old = redis.call('GET', KEYS[2])
if old then redis.call('DEL', ARGV[4] .. ':' .. old) end
redis.call('SET', KEYS[1], ARGV[1], 'EX', ARGV[3])
redis.call('SET', KEYS[2], ARGV[2], 'EX', ARGV[3])
return 1`)

func (r *RedisTokenRepository) StoreExclusiveToken(ctx context.Context, userID int, tokenID string, keyIn string, expiresIn time.Duration) error {
	uid := strconv.Itoa(userID)
	err := storeExclusiveScript.Run(ctx, r.client,
		[]string{r.getRedisKey(keyIn, tokenID), r.getRedisKey(keyIn+"_user", uid)},
		uid, tokenID, ttlSeconds(expiresIn), keyIn).Err()
	if err != nil {
		return fmt.Errorf("could not store exclusive token in redis: %w", err)
	}
	return nil
}

const (
	refreshTokenKey   = "refresh_token"
	refreshFamilyKey  = "refresh_family"
	refreshRotatedKey = "refresh_rotated"
)

// rotateScript exchanges a refresh token for its successor atomically.
// KEYS: 1 presented token, 2 family (holds the current token id), 3 rotated marker of the
// presented token (holds its successor), 4 new token. ARGV: 1 userID, 2 presented id, 3 new id,
// 4 grace seconds, 5 ttl seconds, 6 "1" for a legacy token.
var rotateScript = redis.NewScript(`
local legacy = ARGV[6] == '1'
local current = redis.call('GET', KEYS[2])
if not legacy and not current then return {'invalid', ''} end
local stored = redis.call('GET', KEYS[1])
if not stored then
  if redis.call('EXISTS', KEYS[3]) == 1 then
    redis.call('DEL', KEYS[2])
    return {'reuse', ''}
  end
  return {'invalid', ''}
end
if stored ~= ARGV[1] then return {'invalid', ''} end
local successor = redis.call('GET', KEYS[3])
if successor then return {'replay', successor} end
if current and current ~= ARGV[2] then return {'invalid', ''} end
redis.call('SET', KEYS[3], ARGV[3], 'EX', ARGV[5])
redis.call('EXPIRE', KEYS[1], ARGV[4])
redis.call('SET', KEYS[4], ARGV[1], 'EX', ARGV[5])
redis.call('SET', KEYS[2], ARGV[3], 'EX', ARGV[5])
return {'ok', ''}`)

func (r *RedisTokenRepository) RotateRefreshToken(ctx context.Context, req domain.RotateRequest) (domain.RotateOutcome, string, error) {
	legacy := "0"
	if req.Legacy {
		legacy = "1"
	}
	res, err := rotateScript.Run(ctx, r.client, []string{
		r.getRedisKey(refreshTokenKey, req.PresentedID),
		r.getRedisKey(refreshFamilyKey, req.Family),
		r.getRedisKey(refreshRotatedKey, req.PresentedID),
		r.getRedisKey(refreshTokenKey, req.NewID),
	}, req.UserID, req.PresentedID, req.NewID, ttlSeconds(req.Grace), ttlSeconds(req.TTL), legacy).Slice()
	if err != nil {
		return domain.RotateInvalid, "", fmt.Errorf("could not rotate refresh token in redis: %w", err)
	}
	status, _ := res[0].(string)
	successor, _ := res[1].(string)
	switch status {
	case "ok":
		return domain.RotateOK, "", nil
	case "replay":
		return domain.RotateReplay, successor, nil
	case "reuse":
		return domain.RotateReuse, "", nil
	}
	return domain.RotateInvalid, "", nil
}

// ttlSeconds rounds a lifetime up to whole seconds, at least one (Redis rejects EX 0).
func ttlSeconds(d time.Duration) int64 {
	s := int64((d + time.Second - 1) / time.Second)
	if s < 1 {
		return 1
	}
	return s
}

const sessionsRevokedKey = "sessions_revoked_at"

func (r *RedisTokenRepository) RevokeUserSessions(ctx context.Context, userID int, ttl time.Duration) error {
	key := r.getRedisKey(sessionsRevokedKey, strconv.Itoa(userID))
	if err := r.client.Set(ctx, key, time.Now().Unix(), ttl).Err(); err != nil {
		return fmt.Errorf("could not revoke user sessions in redis: %w", err)
	}
	return nil
}

func (r *RedisTokenRepository) SessionsRevokedAt(ctx context.Context, userID int) (time.Time, error) {
	value, err := r.client.Get(ctx, r.getRedisKey(sessionsRevokedKey, strconv.Itoa(userID))).Int64()
	if err == redis.Nil {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("could not read session revocation from redis: %w", err)
	}
	return time.Unix(value, 0), nil
}

// getRedisKey creates a consistent key for storing refresh tokens in Redis.
func (r *RedisTokenRepository) getRedisKey(key string, tokenID string) string {
	return fmt.Sprintf("%s:%s", key, tokenID)
}
