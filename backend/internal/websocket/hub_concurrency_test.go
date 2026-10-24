package websocket

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestRemoveClientCleansEveryRoomItWasIn(t *testing.T) {
	c, other := newTestClient(1, 4), newTestClient(2, 4)
	hub := startHub(t, c, other)
	hub.AddClientToRoom(c, "a")
	hub.AddClientToRoom(c, "b")
	hub.AddClientToRoom(other, "b")
	if got := len(hub.RoomsOf(c)); got != 2 {
		t.Fatalf("RoomsOf = %d, want 2", got)
	}

	hub.unregister <- c
	time.Sleep(20 * time.Millisecond)

	hub.mu.RLock()
	defer hub.mu.RUnlock()
	if _, ok := hub.rooms["a"]; ok {
		t.Fatal("empty room should be dropped")
	}
	if len(hub.rooms["b"]) != 1 || !hub.rooms["b"][other] {
		t.Fatal("other member must stay in room b")
	}
	if len(hub.users[1]) != 0 {
		t.Fatal("user index should be cleaned")
	}
}

func TestBroadcastNeverBlocksWhenHubIsSaturated(t *testing.T) {
	hub := NewHub() // no Run: nothing drains the channel
	done := make(chan struct{})
	go func() {
		for i := 0; i < broadcastBuffer+50; i++ {
			hub.Broadcast("x", i, "room")
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Broadcast blocked")
	}
	if got := hub.DroppedEvents(); got != 50 {
		t.Fatalf("dropped = %d, want 50", got)
	}
}

// failingBus always errors on publish, like Redis being down.
type failingBus struct{}

func (failingBus) Publish(context.Context, Event) error { return context.DeadlineExceeded }
func (failingBus) Subscribe(context.Context) (<-chan Event, error) {
	return make(chan Event), nil
}

func TestBusFailureFallsBackToLocalDelivery(t *testing.T) {
	c := newTestClient(1, 4)
	hub := startHub(t, c)
	hub.UseBus(failingBus{})
	hub.AddClientToRoom(c, "room")
	hub.Broadcast("ping", 1, "room")
	receive(t, c)
}

func TestShutdownStopsHubAndBusConsumer(t *testing.T) {
	c := newTestClient(1, 4)
	hub := NewHub()
	hub.UseBus(&memBus{})
	runDone := make(chan struct{})
	go func() { hub.Run(); close(runDone) }()
	if err := hub.StartBus(context.Background()); err != nil {
		t.Fatal(err)
	}
	hub.Register(c)

	hub.Shutdown(context.Background())

	select {
	case <-runDone:
	case <-time.After(time.Second):
		t.Fatal("hub loop still running")
	}
	if _, open := <-c.send; open {
		t.Fatal("client should be closed")
	}
	if !hub.isDone() {
		t.Fatal("hub should report done")
	}
	hub.Shutdown(context.Background()) // idempotent
}

func TestHubIsSafeUnderConcurrentUse(t *testing.T) {
	hub := startHub(t)
	var wg sync.WaitGroup
	for u := 0; u < 20; u++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				c := newTestClient(u%5, 2)
				hub.Register(c)
				hub.AddClientToRoom(c, "shared")
				hub.AddClientToRoom(c, "own")
				hub.Broadcast("m", i, "shared")
				hub.SubscribeUser(u%5, "late")
				hub.RoomsOf(c)
				if i%7 == 0 {
					hub.DisconnectUser(u % 5)
				}
				hub.unregister <- c
				// send may already be closed and drained; just consume what is left
				for range c.send {
				}
			}
		}()
	}
	wg.Wait()
}

type sequenceLister struct {
	mu    sync.Mutex
	calls int
	lists [][]string
}

func (l *sequenceLister) GetUserRoomIDs(context.Context, int) ([]string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	i := min(l.calls, len(l.lists)-1)
	l.calls++
	return l.lists[i], nil
}

// A kick landing between the first lookup and the subscription must not leave the socket in the room.
func TestSubscribeDropsRoomsLeftDuringTheLookup(t *testing.T) {
	c := newTestClient(1, 4)
	hub := startHub(t, c)
	lister := &sequenceLister{lists: [][]string{{"kept", "kicked"}, {"kept"}}}

	subscribeToUserRooms(context.Background(), hub, c, lister)

	if !hub.IsClientInRoom(c, "kept") || hub.IsClientInRoom(c, "kicked") {
		t.Fatalf("rooms = %v", hub.RoomsOf(c))
	}
}
