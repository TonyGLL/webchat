package websocket

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"
)

const (
	eventBroadcast   = "broadcast"
	eventSubscribe   = "subscribe"
	eventUnsubscribe = "unsubscribe"
	eventDisconnect  = "disconnect"
	busChannel       = "webchat:ws:events"
)

// Event is what instances exchange so every backend instance can serve its own
// websocket clients: room broadcasts and "subscribe this user to that room" commands.
type Event struct {
	Kind    string          `json:"kind"`
	Type    string          `json:"type,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
	RoomID  string          `json:"room_id"`
	UserID  int             `json:"user_id,omitempty"`
}

// Bus carries events between backend instances.
type Bus interface {
	Publish(ctx context.Context, e Event) error
	// Subscribe returns a stream of every event published by any instance (including this one).
	Subscribe(ctx context.Context) (<-chan Event, error)
}

// RedisBus implements Bus with Redis pub/sub.
type RedisBus struct {
	client *redis.Client
}

func NewRedisBus(client *redis.Client) *RedisBus {
	return &RedisBus{client: client}
}

func (b *RedisBus) Publish(ctx context.Context, e Event) error {
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return b.client.Publish(ctx, busChannel, data).Err()
}

func (b *RedisBus) Subscribe(ctx context.Context) (<-chan Event, error) {
	pubsub := b.client.Subscribe(ctx, busChannel)
	if _, err := pubsub.Receive(ctx); err != nil { // wait for the subscription to be confirmed
		pubsub.Close()
		return nil, err
	}

	events := make(chan Event, 256)
	go func() {
		defer close(events)
		defer pubsub.Close()
		for msg := range pubsub.Channel() {
			var e Event
			if err := json.Unmarshal([]byte(msg.Payload), &e); err != nil {
				log.Printf("websocket bus: ignoring malformed event: %v", err)
				continue
			}
			select {
			case events <- e:
			case <-ctx.Done():
				return
			}
		}
	}()
	return events, nil
}
