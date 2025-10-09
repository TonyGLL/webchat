package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// NewClient establishes a connection to a Redis server using the provided URL.
// It pings the server to verify the connection is active before returning the client.
func NewClient(ctx context.Context, redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("could not parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Ping the Redis server to ensure a connection is established.
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("could not connect to redis: %w", err)
	}

	return client, nil
}