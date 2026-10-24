package websocket

import (
	"context"
	"sync"
	"testing"
	"time"
)

func newTestClient(userID, buffer int) *Client {
	return &Client{UserID: userID, send: make(chan []byte, buffer)}
}

func startHub(t *testing.T, clients ...*Client) *Hub {
	t.Helper()
	hub := NewHub()
	go hub.Run()
	for _, c := range clients {
		hub.Register(c)
	}
	return hub
}

func receive(t *testing.T, c *Client) []byte {
	t.Helper()
	select {
	case msg := <-c.send:
		return msg
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for a message")
		return nil
	}
}

func expectNothing(t *testing.T, c *Client) {
	t.Helper()
	select {
	case msg := <-c.send:
		t.Fatalf("unexpected message %s", msg)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestBroadcastReachesOnlyRoomMembers(t *testing.T) {
	member, outsider := newTestClient(1, 4), newTestClient(2, 4)
	hub := startHub(t, member, outsider)
	hub.AddClientToRoom(member, "room-a")
	hub.AddClientToRoom(outsider, "room-b")

	hub.Broadcast("new_message", map[string]string{"content": "hi"}, "room-a")

	if got := string(receive(t, member)); got != `{"type":"new_message","payload":{"content":"hi"}}` {
		t.Fatalf("unexpected payload: %s", got)
	}
	expectNothing(t, outsider)
}

// A client that never reads used to deadlock the hub (it re-locked a held mutex).
func TestStuckClientDoesNotBlockTheHub(t *testing.T) {
	stuck, healthy := newTestClient(1, 0), newTestClient(2, 4) // stuck has no buffer
	hub := startHub(t, stuck, healthy)
	hub.AddClientToRoom(stuck, "room")
	hub.AddClientToRoom(healthy, "room")

	hub.Broadcast("ping", 1, "room")
	receive(t, healthy)

	// The hub must still serve later broadcasts.
	hub.Broadcast("ping", 2, "room")
	receive(t, healthy)

	select {
	case _, open := <-stuck.send:
		if open {
			t.Fatal("stuck client should have been disconnected")
		}
	case <-time.After(time.Second):
		t.Fatal("stuck client was not disconnected")
	}
}

func TestSubscribeUserAddsAllTheirConnections(t *testing.T) {
	a1, a2, b := newTestClient(1, 4), newTestClient(1, 4), newTestClient(2, 4)
	hub := startHub(t, a1, a2, b)

	hub.SubscribeUser(1, "new-room")
	hub.Broadcast("new_message", 1, "new-room")

	receive(t, a1)
	receive(t, a2)
	expectNothing(t, b)
}

func TestDisconnectedClientIsNotAddedToRooms(t *testing.T) {
	c := newTestClient(1, 4)
	hub := startHub(t, c)
	hub.unregister <- c
	time.Sleep(20 * time.Millisecond)

	hub.AddClientToRoom(c, "room")

	hub.mu.Lock()
	defer hub.mu.Unlock()
	if len(hub.rooms) != 0 {
		t.Fatal("a disconnected client must not be added to a room")
	}
}

// memBus is an in-memory Bus shared by several hubs, simulating Redis pub/sub between instances.
type memBus struct {
	mu   sync.Mutex
	subs []chan Event
}

func (b *memBus) Publish(_ context.Context, e Event) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, s := range b.subs {
		s <- e
	}
	return nil
}

func (b *memBus) Subscribe(context.Context) (<-chan Event, error) {
	ch := make(chan Event, 16)
	b.mu.Lock()
	b.subs = append(b.subs, ch)
	b.mu.Unlock()
	return ch, nil
}

func TestEventsCrossInstancesThroughTheBus(t *testing.T) {
	bus := &memBus{}
	newInstance := func(clients ...*Client) *Hub {
		hub := startHub(t, clients...)
		hub.UseBus(bus)
		if err := hub.StartBus(context.Background()); err != nil {
			t.Fatal(err)
		}
		return hub
	}
	onA, onB := newTestClient(1, 4), newTestClient(2, 4)
	instanceA := newInstance(onA)
	instanceB := newInstance(onB)

	// user 2 is connected to instance B only; a command issued on A must still reach them
	instanceA.SubscribeUser(2, "room")
	time.Sleep(50 * time.Millisecond)
	instanceA.AddClientToRoom(onA, "room")

	// a broadcast issued on B reaches the client connected to A, and the one on B
	instanceB.Broadcast("new_message", map[string]int{"n": 1}, "room")

	want := `{"type":"new_message","payload":{"n":1}}`
	if got := string(receive(t, onA)); got != want {
		t.Fatalf("instance A client got %s", got)
	}
	if got := string(receive(t, onB)); got != want {
		t.Fatalf("instance B client got %s", got)
	}
}
