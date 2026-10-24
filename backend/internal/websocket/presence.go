package websocket

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	presenceKey = "webchat:presence"
	// presenceTTL is how long a connection counts as online without a heartbeat. Heartbeats
	// come with every pong (about every 54 s), so a crashed instance's users expire by themselves.
	presenceTTL = 2 * time.Minute
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

// RedisPresence keeps a sorted set whose members are "userID:clientID" and whose scores are
// the time each connection expires.
type RedisPresence struct {
	client *redis.Client
	now    func() time.Time
}

func NewRedisPresence(client *redis.Client) *RedisPresence {
	return &RedisPresence{client: client, now: time.Now}
}

func (p *RedisPresence) member(userID int, clientID string) string {
	return strconv.Itoa(userID) + ":" + clientID
}

func (p *RedisPresence) expiry() float64 {
	return float64(p.now().Add(presenceTTL).Unix())
}

// liveMembers returns the non-expired members, after dropping the expired ones.
func (p *RedisPresence) liveMembers(ctx context.Context) ([]string, error) {
	now := strconv.FormatInt(p.now().Unix(), 10)
	if err := p.client.ZRemRangeByScore(ctx, presenceKey, "-inf", now).Err(); err != nil {
		return nil, err
	}
	return p.client.ZRangeByScore(ctx, presenceKey, &redis.ZRangeBy{Min: "(" + now, Max: "+inf"}).Result()
}

func connectionsOf(members []string, userID int) int {
	prefix := strconv.Itoa(userID) + ":"
	n := 0
	for _, m := range members {
		if strings.HasPrefix(m, prefix) {
			n++
		}
	}
	return n
}

func (p *RedisPresence) Connect(ctx context.Context, userID int, clientID string) (bool, error) {
	if err := p.client.ZAdd(ctx, presenceKey, &redis.Z{Score: p.expiry(), Member: p.member(userID, clientID)}).Err(); err != nil {
		return false, err
	}
	members, err := p.liveMembers(ctx)
	if err != nil {
		return false, err
	}
	return connectionsOf(members, userID) == 1, nil
}

func (p *RedisPresence) Touch(ctx context.Context, userID int, clientID string) error {
	return p.client.ZAdd(ctx, presenceKey, &redis.Z{Score: p.expiry(), Member: p.member(userID, clientID)}).Err()
}

func (p *RedisPresence) Disconnect(ctx context.Context, userID int, clientID string) (bool, error) {
	if err := p.client.ZRem(ctx, presenceKey, p.member(userID, clientID)).Err(); err != nil {
		return false, err
	}
	members, err := p.liveMembers(ctx)
	if err != nil {
		return false, err
	}
	return connectionsOf(members, userID) == 0, nil
}

func (p *RedisPresence) OnlineUsers(ctx context.Context, userIDs []int) (map[int]bool, error) {
	members, err := p.liveMembers(ctx)
	if err != nil {
		return nil, err
	}
	online := make(map[int]bool, len(userIDs))
	wanted := make(map[int]bool, len(userIDs))
	for _, id := range userIDs {
		wanted[id] = true
	}
	for _, m := range members {
		idPart, _, _ := strings.Cut(m, ":")
		if id, err := strconv.Atoi(idPart); err == nil && wanted[id] {
			online[id] = true
		}
	}
	return online, nil
}
