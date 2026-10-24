package websocket

import (
	"context"
	"sort"
	"strconv"
	"strings"
)

// OnlineUserIDs returns up to max distinct online user ids (lowest ids first, so the result
// is stable between calls). Used to surface online people in a search without scanning users.
func (p *RedisPresence) OnlineUserIDs(ctx context.Context, max int) ([]int, error) {
	members, err := p.liveMembers(ctx)
	if err != nil {
		return nil, err
	}
	seen := make(map[int]bool)
	for _, m := range members {
		idPart, _, _ := strings.Cut(m, ":")
		if id, err := strconv.Atoi(idPart); err == nil {
			seen[id] = true
		}
	}
	ids := make([]int, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	if max > 0 && len(ids) > max {
		ids = ids[:max]
	}
	return ids, nil
}
