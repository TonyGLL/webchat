package websocket

import (
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
