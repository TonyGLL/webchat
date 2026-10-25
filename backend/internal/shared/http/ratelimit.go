package http

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Counter counts hits for a key inside a fixed window.
type Counter interface {
	// Hit registers one hit and returns the total in the current window and the time left in it.
	Hit(ctx context.Context, key string, window time.Duration) (count int64, ttl time.Duration, err error)
}

// RedisCounter is a Counter backed by Redis (INCR + EXPIRE).
type RedisCounter struct {
	client *redis.Client
}

func NewRedisCounter(client *redis.Client) *RedisCounter {
	return &RedisCounter{client: client}
}

func (r *RedisCounter) Hit(ctx context.Context, key string, window time.Duration) (int64, time.Duration, error) {
	pipe := r.client.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.ExpireNX(ctx, key, window)
	ttl := pipe.PTTL(ctx, key)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, 0, err
	}
	return incr.Val(), ttl.Val(), nil
}

// Rule limits a route: at most Limit requests per client IP every Window.
type Rule struct {
	Method string
	Path   string // route pattern as registered in Gin, e.g. "/api/v1/auth/login"
	Limit  int64
	Window time.Duration
}

// RateLimit returns a middleware that applies the first rule matching the request's
// method and route. If the counter is unavailable it lets the request through
// (availability over strictness) and logs the problem.
func RateLimit(counter Counter, rules []Rule) gin.HandlerFunc {
	type key struct{ method, path string }
	byRoute := make(map[key]Rule, len(rules))
	for _, r := range rules {
		byRoute[key{r.Method, r.Path}] = r
	}

	return func(c *gin.Context) {
		rule, ok := byRoute[key{c.Request.Method, c.FullPath()}]
		if !ok {
			c.Next()
			return
		}

		redisKey := "rl:" + rule.Method + ":" + rule.Path + ":" + c.ClientIP()
		count, ttl, err := counter.Hit(c.Request.Context(), redisKey, rule.Window)
		if err != nil {
			log.Printf("rate limiter unavailable: %v", err)
			c.Next()
			return
		}

		if count > rule.Limit {
			retryAfter := int(ttl.Seconds()) + 1
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests, try again later"})
			return
		}
		c.Next()
	}
}
