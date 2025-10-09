package persistence

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"backend/internal/auth/domain"
	shared_domain "backend/internal/shared/domain"

	"github.com/go-redis/redis/v8"
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

// StoreRefreshToken stores the refresh token's ID in Redis with the user ID as the value.
// The key is prefixed to avoid collisions and for better organization.
func (r *RedisTokenRepository) StoreRefreshToken(ctx context.Context, userID int, tokenID string, expiresIn time.Duration) error {
	key := r.getRedisKey(tokenID)
	err := r.client.Set(ctx, key, userID, expiresIn).Err()
	if err != nil {
		return fmt.Errorf("could not store refresh token in redis: %w", err)
	}
	return nil
}

// GetUserIDByRefreshToken retrieves the user ID associated with the given refresh token ID.
// If the token is not found in Redis, it returns a domain-specific ErrNotFound.
func (r *RedisTokenRepository) GetUserIDByRefreshToken(ctx context.Context, tokenID string) (int, error) {
	key := r.getRedisKey(tokenID)
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
func (r *RedisTokenRepository) DeleteRefreshToken(ctx context.Context, tokenID string) error {
	key := r.getRedisKey(tokenID)
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("could not delete refresh token from redis: %w", err)
	}
	return nil
}

// getRedisKey creates a consistent key for storing refresh tokens in Redis.
func (r *RedisTokenRepository) getRedisKey(tokenID string) string {
	return fmt.Sprintf("refresh_token:%s", tokenID)
}