package websocket

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	// legacyPresenceKey was the single sorted set holding every connection. It is no longer
	// read; NewRedisPresence deletes it once so it does not linger without a TTL.
	legacyPresenceKey = "webchat:presence"
	// presenceKeyPrefix prefixes one sorted set per user ("webchat:presence:u:<id>") whose members
	// are client IDs and whose scores are the time each connection expires.
	presenceKeyPrefix = "webchat:presence:u:"
	// presenceTTL is how long a connection counts as online without a heartbeat. Heartbeats
	// come with every pong (about every 54 s), so a crashed instance's users expire by themselves.
	presenceTTL = 2 * time.Minute
	// maxOnlineLookup caps how many users one OnlineUsers call checks, and onlineChunk how many
	// checks go in one pipeline.
	maxOnlineLookup = 1000
	onlineChunk     = 200
)

// PresenceStore tracks which users have at least one open connection, across all instances.
type PresenceStore interface {
	// Connect registers a connection and reports whether it is the user's first one.
	Connect(ctx context.Context, userID int, clientID string) (first bool, err error)
	// Touch renews a connection's heartbeat.
	Touch(ctx context.Context, userID int, clientID string) error
	// Disconnect removes a connection and reports whether it was the user's last one.
	Disconnect(ctx context.Context, userID int, clientID string) (last bool, err error)
	// OnlineUsers returns which of the given users are online.
	OnlineUsers(ctx context.Context, userIDs []int) (map[int]bool, error)
}

// RedisPresence keeps, per user, a sorted set of client IDs scored by the time each connection
// expires. Every decision touches only that user's key, so its cost does not grow with the
// number of connected users.
type RedisPresence struct {
	client *redis.Client
	now    func() time.Time
}

// NewRedisPresence builds the store and, in the background, drops the legacy global key.
func NewRedisPresence(client *redis.Client) *RedisPresence {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), redisOpTimeout)
		defer cancel()
		_ = client.Del(ctx, legacyPresenceKey).Err()
	}()
	return &RedisPresence{client: client, now: time.Now}
}

func userKey(userID int) string {
	return presenceKeyPrefix + strconv.Itoa(userID)
}

func (p *RedisPresence) expiry() float64 {
	return float64(p.now().Add(presenceTTL).Unix())
}

// connectScript registers a connection, prunes the user's expired ones and returns how many
// are left, atomically. KEYS[1]=user key; ARGV: now, expiry, client id, key TTL in seconds.
var connectScript = redis.NewScript(`
redis.call('ZADD', KEYS[1], ARGV[2], ARGV[3])
redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', ARGV[1])
redis.call('EXPIRE', KEYS[1], ARGV[4])
return redis.call('ZCARD', KEYS[1])
`)

// disconnectScript removes a connection, prunes expired ones and returns how many are left.
var disconnectScript = redis.NewScript(`
redis.call('ZREM', KEYS[1], ARGV[2])
redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', ARGV[1])
return redis.call('ZCARD', KEYS[1])
`)

func (p *RedisPresence) Connect(ctx context.Context, userID int, clientID string) (bool, error) {
	n, err := connectScript.Run(ctx, p.client, []string{userKey(userID)},
		p.now().Unix(), p.expiry(), clientID, int(presenceTTL.Seconds())).Int()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

func (p *RedisPresence) Touch(ctx context.Context, userID int, clientID string) error {
	key := userKey(userID)
	_, err := p.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.ZAdd(ctx, key, &redis.Z{Score: p.expiry(), Member: clientID})
		pipe.Expire(ctx, key, presenceTTL)
		return nil
	})
	return err
}

func (p *RedisPresence) Disconnect(ctx context.Context, userID int, clientID string) (bool, error) {
	n, err := disconnectScript.Run(ctx, p.client, []string{userKey(userID)}, p.now().Unix(), clientID).Int()
	if err != nil {
		return false, err
	}
	return n == 0, nil
}

// OnlineUsers checks each requested user with one ZCOUNT of their live connections, pipelined.
// Lookups are capped (extra ids are reported offline) so a huge listing cannot flood Redis.
func (p *RedisPresence) OnlineUsers(ctx context.Context, userIDs []int) (map[int]bool, error) {
	if len(userIDs) > maxOnlineLookup {
		userIDs = userIDs[:maxOnlineLookup]
	}
	online := make(map[int]bool, len(userIDs))
	after := "(" + strconv.FormatInt(p.now().Unix(), 10)
	for start := 0; start < len(userIDs); start += onlineChunk {
		chunk := userIDs[start:min(start+onlineChunk, len(userIDs))]
		pipe := p.client.Pipeline()
		counts := make([]*redis.IntCmd, len(chunk))
		for i, id := range chunk {
			counts[i] = pipe.ZCount(ctx, userKey(id), after, "+inf")
		}
		if _, err := pipe.Exec(ctx); err != nil {
			return nil, err
		}
		for i, id := range chunk {
			if counts[i].Val() > 0 {
				online[id] = true
			}
		}
	}
	return online, nil
}
