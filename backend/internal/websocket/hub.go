// Package websocket handles real-time communication using WebSockets.
package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
)

// broadcastBuffer lets request handlers publish events without waiting for the hub loop.
const broadcastBuffer = 256

// Hub maintains the set of active clients and broadcasts messages to the
// clients in specific rooms.
type Hub struct {
	// Inbound messages from the application.
	broadcast chan *Message

	// Unregister requests from clients.
	unregister chan *Client

	// bus, when set, relays events between backend instances (see UseBus).
	bus Bus

	// mu protects clients and rooms, which are touched by the hub loop and by other goroutines.
	mu sync.Mutex

	// Registered clients.
	clients map[*Client]bool

	// A map where keys are room IDs and values are the clients subscribed to that room.
	rooms map[string]map[*Client]bool
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan *Message, broadcastBuffer),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
	}
}

// Run starts the hub's event loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.unregister:
			h.mu.Lock()
			h.removeClientLocked(client)
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.handleBroadcast(message)
		}
	}
}

// Register adds a client. It is synchronous so the caller can subscribe the client
// to rooms right after, knowing it is already known to the hub.
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client] = true
}

// removeClientLocked disconnects a client everywhere. The caller must hold h.mu.
func (h *Hub) removeClientLocked(client *Client) {
	if !h.clients[client] {
		return
	}
	delete(h.clients, client)
	close(client.send)
	for roomID, members := range h.rooms {
		delete(members, client)
		if len(members) == 0 {
			delete(h.rooms, roomID)
		}
	}
}

// handleBroadcast sends a message to all clients in the specified room.
// Clients whose send buffer is full are considered stuck and are dropped.
func (h *Hub) handleBroadcast(message *Message) {
	payloadBytes, err := json.Marshal(message.Payload)
	if err != nil {
		log.Printf("error marshalling broadcast message payload: %v", err)
		return
	}
	finalBytes, err := json.Marshal(WebsocketMessage{Type: message.Type, Payload: json.RawMessage(payloadBytes)})
	if err != nil {
		log.Printf("error marshalling final websocket message: %v", err)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	var stuck []*Client
	for client := range h.rooms[message.RoomID] {
		select {
		case client.send <- finalBytes:
		default:
			stuck = append(stuck, client)
		}
	}
	for _, client := range stuck {
		h.removeClientLocked(client)
	}
}

// IsClientInRoom reports whether the client is subscribed to the room.
func (h *Hub) IsClientInRoom(client *Client, roomID string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.rooms[roomID][client]
}

// RoomsOf lists the rooms a client is subscribed to.
func (h *Hub) RoomsOf(client *Client) []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	var rooms []string
	for roomID, members := range h.rooms {
		if members[client] {
			rooms = append(rooms, roomID)
		}
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

func (h *Hub) addToRoomLocked(client *Client, roomID string) {
	if _, ok := h.rooms[roomID]; !ok {
		h.rooms[roomID] = make(map[*Client]bool)
	}
	h.rooms[roomID][client] = true
}

// SubscribeUser subscribes every open connection of a user to a room, so rooms
// created or joined after connecting start receiving events immediately. With a bus
// configured the command reaches every instance, since the user's connections may be anywhere.
func (h *Hub) SubscribeUser(userID int, roomID string) {
	if h.bus != nil {
		err := h.bus.Publish(context.Background(), Event{Kind: eventSubscribe, RoomID: roomID, UserID: userID})
		if err == nil {
			return
		}
		log.Printf("websocket bus unavailable, subscribing locally: %v", err)
	}
	h.subscribeUserLocal(userID, roomID)
}

// UnsubscribeUser removes every open connection of a user from a room (for example after leaving it).
func (h *Hub) UnsubscribeUser(userID int, roomID string) {
	if h.bus != nil {
		err := h.bus.Publish(context.Background(), Event{Kind: eventUnsubscribe, RoomID: roomID, UserID: userID})
		if err == nil {
			return
		}
		log.Printf("websocket bus unavailable, unsubscribing locally: %v", err)
	}
	h.unsubscribeUserLocal(userID, roomID)
}

// DisconnectUser closes every connection of a user (for example after deleting their account).
func (h *Hub) DisconnectUser(userID int) {
	if h.bus != nil {
		err := h.bus.Publish(context.Background(), Event{Kind: eventDisconnect, UserID: userID})
		if err == nil {
			return
		}
		log.Printf("websocket bus unavailable, disconnecting locally: %v", err)
	}
	h.disconnectUserLocal(userID)
}

func (h *Hub) disconnectUserLocal(userID int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for client := range h.clients {
		if client.UserID == userID {
			h.removeClientLocked(client)
		}
	}
}

func (h *Hub) unsubscribeUserLocal(userID int, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	members := h.rooms[roomID]
	for client := range members {
		if client.UserID == userID {
			delete(members, client)
		}
	}
	if len(members) == 0 {
		delete(h.rooms, roomID)
	}
}

func (h *Hub) subscribeUserLocal(userID int, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		if client.UserID == userID {
			h.addToRoomLocked(client, roomID)
		}
	}
}

// UseBus makes the hub share events with other backend instances. Call before StartBus.
func (h *Hub) UseBus(bus Bus) { h.bus = bus }

// StartBus consumes events from the bus until ctx is cancelled. Without a bus it does nothing.
func (h *Hub) StartBus(ctx context.Context) error {
	if h.bus == nil {
		return nil
	}
	events, err := h.bus.Subscribe(ctx)
	if err != nil {
		return err
	}
	go func() {
		for e := range events {
			switch e.Kind {
			case eventBroadcast:
				h.broadcast <- &Message{Type: e.Type, Payload: e.Payload, RoomID: e.RoomID}
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

// Broadcast publishes an event for the clients of a room. With a bus it goes through it
// (so clients on other instances get it too); if the bus fails it is delivered locally.
func (h *Hub) Broadcast(messageType string, payload interface{}, roomID string) {
	if h.bus != nil {
		data, err := json.Marshal(payload)
		if err == nil {
			err = h.bus.Publish(context.Background(), Event{Kind: eventBroadcast, Type: messageType, Payload: data, RoomID: roomID})
		}
		if err == nil {
			return
		}
		log.Printf("websocket bus unavailable, broadcasting locally: %v", err)
	}
	h.broadcast <- &Message{Type: messageType, Payload: payload, RoomID: roomID}
}
