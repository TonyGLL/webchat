package websocket

import (
	"context"
	"sort"
	"strconv"

	"github.com/go-redis/redis/v8"
)

// OnlineUserIDs returns up to max online user ids, lowest first within the window it reads.
// It reads the online index with a LIMIT, so its cost does not depend on how many users are
// connected; users whose heartbeats stopped (a crashed instance) are pruned on the way.
func (p *RedisPresence) OnlineUserIDs(ctx context.Context, max int) ([]int, error) {
	if max <= 0 || max > maxOnlineLookup {
		max = maxOnlineLookup
	}
	now := strconv.FormatInt(p.now().Unix(), 10)
	pipe := p.client.TxPipeline()
	pipe.ZRemRangeByScore(ctx, onlineIndexKey, "-inf", now)
	members := pipe.ZRangeByScore(ctx, onlineIndexKey, &redis.ZRangeBy{Min: "(" + now, Max: "+inf", Count: int64(max)})
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}
	ids := make([]int, 0, max)
	for _, m := range members.Val() {
		if id, err := strconv.Atoi(m); err == nil {
			ids = append(ids, id)
		}
	}
	sort.Ints(ids)
	return ids, nil
}
