// Package websocket handles real-time communication using WebSockets.
package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// broadcastBuffer lets request handlers publish events without waiting for the hub loop.
const broadcastBuffer = 256

// busTimeout bounds every publish to the bus so a stalled Redis never blocks a request handler.
const busTimeout = 2 * time.Second

// logEvery is the minimum gap between two lines of the same noisy log (bus down, dropped events).
const logEvery = 10 * time.Second

// limitedLog prints at most one line per interval and counts the ones it swallowed, so a Redis
// outage does not turn every broadcast into a log line.
type limitedLog struct {
	mu         sync.Mutex
	last       time.Time
	suppressed int
}

func (l *limitedLog) Printf(format string, args ...any) {
	l.mu.Lock()
	now := time.Now()
	if !l.last.IsZero() && now.Sub(l.last) < logEvery {
		l.suppressed++
		l.mu.Unlock()
		return
	}
	skipped := l.suppressed
	l.last, l.suppressed = now, 0
	l.mu.Unlock()
	if skipped > 0 {
		format += " (+%d similar suppressed)"
		args = append(args, skipped)
	}
	log.Printf(format, args...)
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients in specific rooms.
type Hub struct {
	// Inbound messages from the application.
	broadcast chan *Message

	// Unregister requests from clients.
	unregister chan *Client

	// bus, when set, relays events between backend instances (see UseBus).
	bus Bus

	// done is closed by Shutdown; it stops the hub loop and the bus consumer.
	done     chan struct{}
	stopOnce sync.Once
	// busCancel stops the bus consumer started by StartBus.
	busCancel context.CancelFunc

	// dropped counts broadcasts discarded because the hub could not keep up.
	dropped atomic.Uint64
	busLog  limitedLog
	dropLog limitedLog

	// mu protects clients, users and rooms (and each client's rooms set), which are touched by
	// the hub loop and by other goroutines. Fan-out only takes the read lock to copy a room.
	mu sync.RWMutex

	// Registered clients.
	clients map[*Client]bool

	// Registered clients by user, so per-user commands do not scan every connection.
	users map[int]map[*Client]struct{}

	// A map where keys are room IDs and values are the clients subscribed to that room.
	rooms map[string]map[*Client]bool

	// maxConnsPerUser caps open sockets per user; the oldest is closed to make room.
	maxConnsPerUser int
	// seq orders registrations so the oldest connection of a user can be found.
	seq uint64
}

// DefaultMaxConnsPerUser is how many sockets one account may keep open (several tabs and devices).
const DefaultMaxConnsPerUser = 8

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan *Message, broadcastBuffer),
		unregister: make(chan *Client),
		done:       make(chan struct{}),
		clients:    make(map[*Client]bool),
		users:      make(map[int]map[*Client]struct{}),
		rooms:      make(map[string]map[*Client]bool),

		maxConnsPerUser: DefaultMaxConnsPerUser,
	}
}

// Run starts the hub's event loop. It returns after Shutdown.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.unregister:
			h.mu.Lock()
			h.removeClientLocked(client)
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.handleBroadcast(message)
		case <-h.done:
			return
		}
	}
}

// Shutdown tells every client the server is restarting (close code 1012) and stops the hub
// loop and the bus consumer. It waits for the close frames until ctx expires.
func (h *Hub) Shutdown(ctx context.Context) {
	h.stopOnce.Do(func() {
		close(h.done)
		if h.busCancel != nil {
			h.busCancel()
		}
	})

	h.mu.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	deadline := time.Now().Add(writeWait)
	if d, ok := ctx.Deadline(); ok {
		deadline = d
	}
	var wg sync.WaitGroup
	for _, c := range clients {
		if c.conn == nil {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := websocket.FormatCloseMessage(websocket.CloseServiceRestart, "server restarting")
			_ = c.conn.WriteControl(websocket.CloseMessage, msg, deadline)
		}()
	}
	finished := make(chan struct{})
	go func() { wg.Wait(); close(finished) }()
	select {
	case <-finished:
	case <-ctx.Done():
	}

	h.mu.Lock()
	for _, c := range clients {
		h.removeClientLocked(c)
	}
	h.mu.Unlock()
}

// isDone reports whether Shutdown was called.
func (h *Hub) isDone() bool {
	select {
	case <-h.done:
		return true
	default:
		return false
	}
}

// isRegistered reports whether the client is still known to the hub.
func (h *Hub) isRegistered(client *Client) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients[client]
}

// Register adds a client. It is synchronous so the caller can subscribe the client
// to rooms right after, knowing it is already known to the hub.
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.seq++
	client.seq = h.seq
	// Make room for the new connection by closing the user's oldest ones.
	for limit := h.maxConnsPerUser; limit > 0 && len(h.users[client.UserID]) >= limit; {
		var oldest *Client
		for c := range h.users[client.UserID] {
			if oldest == nil || c.seq < oldest.seq {
				oldest = c
			}
		}
		h.removeClientLocked(oldest)
	}
	h.clients[client] = true
	if h.users[client.UserID] == nil {
		h.users[client.UserID] = make(map[*Client]struct{})
	}
	h.users[client.UserID][client] = struct{}{}
}

// removeClientLocked disconnects a client everywhere, in O(rooms of the client).
// The caller must hold h.mu for writing.
func (h *Hub) removeClientLocked(client *Client) {
	if !h.clients[client] {
		return
	}
	delete(h.clients, client)
	if set := h.users[client.UserID]; set != nil {
		delete(set, client)
		if len(set) == 0 {
			delete(h.users, client.UserID)
		}
	}
	client.closeSend()
	for roomID := range client.rooms {
		h.leaveRoomLocked(client, roomID)
	}
	client.rooms = nil
}

// leaveRoomLocked removes a client from one room, dropping the room when it empties.
func (h *Hub) leaveRoomLocked(client *Client, roomID string) {
	members := h.rooms[roomID]
	delete(members, client)
	if len(members) == 0 {
		delete(h.rooms, roomID)
	}
	delete(client.rooms, roomID)
}

// handleBroadcast sends a message to all clients in the specified room.
// Clients whose send buffer is full are considered stuck and are dropped.
// The envelope is marshalled once and delivered outside the lock.
func (h *Hub) handleBroadcast(message *Message) {
	// Events relayed by the bus already carry their payload as JSON.
	payload, ok := message.Payload.(json.RawMessage)
	if !ok {
		var err error
		payload, err = json.Marshal(message.Payload)
		if err != nil {
			log.Printf("error marshalling broadcast message payload: %v", err)
			return
		}
	}
	finalBytes, err := json.Marshal(WebsocketMessage{Type: message.Type, Payload: payload})
	if err != nil {
		log.Printf("error marshalling final websocket message: %v", err)
		return
	}

	h.mu.RLock()
	members := h.rooms[message.RoomID]
	targets := make([]*Client, 0, len(members))
	for client := range members {
		targets = append(targets, client)
	}
	h.mu.RUnlock()

	var stuck []*Client
	for _, client := range targets {
		if !client.trySend(finalBytes) {
			stuck = append(stuck, client)
		}
	}
	if len(stuck) == 0 {
		return
	}
	h.mu.Lock()
	for _, client := range stuck {
		h.removeClientLocked(client)
	}
	h.mu.Unlock()
}

// IsClientInRoom reports whether the client is subscribed to the room.
func (h *Hub) IsClientInRoom(client *Client, roomID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.rooms[roomID][client]
}

// RoomsOf lists the rooms a client is subscribed to.
func (h *Hub) RoomsOf(client *Client) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	rooms := make([]string, 0, len(client.rooms))
	for roomID := range client.rooms {
		rooms = append(rooms, roomID)
	}
	return rooms
}

// AddClientToRoom subscribes a connected client to a room.
func (h *Hub) AddClientToRoom(client *Client, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.clients[client] {
		return // already disconnected: do not leak it into a room
	}
	h.addToRoomLocked(client, roomID)
}

// RemoveClientFromRoom unsubscribes one connection from a room.
func (h *Hub) RemoveClientFromRoom(client *Client, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[roomID][client] {
		h.leaveRoomLocked(client, roomID)
	}
}

func (h *Hub) addToRoomLocked(client *Client, roomID string) {
	if _, ok := h.rooms[roomID]; !ok {
		h.rooms[roomID] = make(map[*Client]bool)
	}
	h.rooms[roomID][client] = true
	if client.rooms == nil {
		client.rooms = make(map[string]struct{})
	}
	client.rooms[roomID] = struct{}{}
}

// SubscribeUser subscribes every open connection of a user to a room, so rooms
// created or joined after connecting start receiving events immediately. With a bus
// configured the command reaches every instance, since the user's connections may be anywhere.
func (h *Hub) SubscribeUser(userID int, roomID string) {
	if h.bus != nil {
		err := h.publish(Event{Kind: eventSubscribe, RoomID: roomID, UserID: userID})
		if err == nil {
			return
		}
		h.busLog.Printf("websocket bus unavailable, subscribing locally: %v", err)
	}
	h.subscribeUserLocal(userID, roomID)
}

// UnsubscribeUser removes every open connection of a user from a room (for example after leaving it).
func (h *Hub) UnsubscribeUser(userID int, roomID string) {
	if h.bus != nil {
		err := h.publish(Event{Kind: eventUnsubscribe, RoomID: roomID, UserID: userID})
		if err == nil {
			return
		}
		h.busLog.Printf("websocket bus unavailable, unsubscribing locally: %v", err)
	}
	h.unsubscribeUserLocal(userID, roomID)
}

// DisconnectUser closes every connection of a user (for example after deleting their account).
func (h *Hub) DisconnectUser(userID int) {
	if h.bus != nil {
		err := h.publish(Event{Kind: eventDisconnect, UserID: userID})
		if err == nil {
			return
		}
		h.busLog.Printf("websocket bus unavailable, disconnecting locally: %v", err)
	}
	h.disconnectUserLocal(userID)
}

func (h *Hub) disconnectUserLocal(userID int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for client := range h.users[userID] {
		h.removeClientLocked(client) // deleting from the set while ranging is safe in Go
	}
}

func (h *Hub) unsubscribeUserLocal(userID int, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for client := range h.users[userID] {
		if h.rooms[roomID][client] {
			h.leaveRoomLocked(client, roomID)
		}
	}
}

func (h *Hub) subscribeUserLocal(userID int, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for client := range h.users[userID] {
		h.addToRoomLocked(client, roomID)
	}
}

// UseBus makes the hub share events with other backend instances. Call before StartBus.
func (h *Hub) UseBus(bus Bus) { h.bus = bus }

// publish sends an event to the bus with a short timeout.
func (h *Hub) publish(e Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), busTimeout)
	defer cancel()
	return h.bus.Publish(ctx, e)
}

// StartBus consumes events from the bus until ctx is cancelled or the hub is shut down.
// Without a bus it does nothing.
func (h *Hub) StartBus(ctx context.Context) error {
	if h.bus == nil {
		return nil
	}
	ctx, h.busCancel = context.WithCancel(ctx)
	events, err := h.bus.Subscribe(ctx)
	if err != nil {
		h.busCancel()
		return err
	}
	go func() {
		for {
			var e Event
			select {
			case ev, ok := <-events:
				if !ok {
					return
				}
				e = ev
			case <-ctx.Done():
				return
			}
			switch e.Kind {
			case eventBroadcast:
				h.enqueue(&Message{Type: e.Type, Payload: e.Payload, RoomID: e.RoomID})
			case eventSubscribe:
				h.subscribeUserLocal(e.UserID, e.RoomID)
			case eventDisconnect:
				h.disconnectUserLocal(e.UserID)
			case eventUnsubscribe:
				h.unsubscribeUserLocal(e.UserID, e.RoomID)
			}
		}
	}()
	return nil
}

// enqueue hands a message to the hub loop without ever blocking the caller: if the loop is
// saturated (or stopped) the event is dropped and counted, which beats stalling request handlers.
func (h *Hub) enqueue(m *Message) {
	select {
	case h.broadcast <- m:
	default:
		n := h.dropped.Add(1)
		h.dropLog.Printf("websocket hub saturated, dropping broadcast events (%d dropped so far)", n)
	}
}

// DroppedEvents returns how many broadcasts were discarded because the hub could not keep up.
func (h *Hub) DroppedEvents() uint64 { return h.dropped.Load() }

// Broadcast publishes an event for the clients of a room. With a bus it goes through it
// (so clients on other instances get it too); if the bus fails it is delivered locally.
// It never blocks the caller.
func (h *Hub) Broadcast(messageType string, payload interface{}, roomID string) {
	if h.bus != nil {
		data, err := json.Marshal(payload)
		if err == nil {
			err = h.publish(Event{Kind: eventBroadcast, Type: messageType, Payload: data, RoomID: roomID})
		}
		if err == nil {
			return
		}
		h.busLog.Printf("websocket bus unavailable, broadcasting locally: %v", err)
	}
	h.enqueue(&Message{Type: messageType, Payload: payload, RoomID: roomID})
}
