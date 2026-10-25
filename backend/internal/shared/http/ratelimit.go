package http

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/netip"
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

// RedisCounter is a Counter backed by Redis.
type RedisCounter struct {
	client *redis.Client
}

func NewRedisCounter(client *redis.Client) *RedisCounter {
	return &RedisCounter{client: client}
}

// hitScript increments the counter and arms the expiry on the first hit. EXPIRE ... NX would do
// it in one command but only exists from Redis 7.0; this works on 6.x and later (and the key can
// never be left without a TTL, even if a previous client crashed between INCR and EXPIRE).
var hitScript = redis.NewScript(`
local n = redis.call('INCR', KEYS[1])
local ttl = redis.call('PTTL', KEYS[1])
if ttl < 0 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
  ttl = tonumber(ARGV[1])
end
return {n, ttl}
`)

func (r *RedisCounter) Hit(ctx context.Context, key string, window time.Duration) (int64, time.Duration, error) {
	res, err := hitScript.Run(ctx, r.client, []string{key}, window.Milliseconds()).Int64Slice()
	if err != nil {
		return 0, 0, err
	}
	if len(res) != 2 {
		return 0, 0, errors.New("rate limiter: unexpected script reply")
	}
	return res[0], time.Duration(res[1]) * time.Millisecond, nil
}

// limiterTimeout bounds one counter call so a stalled Redis cannot hold every request.
const limiterTimeout = 300 * time.Millisecond

// clientKey is the address a rate-limit bucket is keyed on. IPv6 clients are grouped by their /64
// prefix (the smallest allocation an ISP or host normally hands out), so rotating addresses inside
// one allocation does not give an attacker a fresh bucket each time. IPv4 is used as is.
func clientKey(ip string) string {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return ip
	}
	addr = addr.Unmap()
	if addr.Is6() {
		if p, err := addr.Prefix(64); err == nil {
			return p.Masked().String()
		}
	}
	return addr.String()
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

		redisKey := "rl:" + rule.Method + ":" + rule.Path + ":" + clientKey(c.ClientIP())
		ctx, cancel := context.WithTimeout(c.Request.Context(), limiterTimeout)
		count, ttl, err := counter.Hit(ctx, redisKey, rule.Window)
		cancel()
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
