// Package websocket handles real-time communication using WebSockets.
package websocket

import (
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
// created or joined after connecting start receiving events immediately.
func (h *Hub) SubscribeUser(userID int, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		if client.UserID == userID {
			h.addToRoomLocked(client, roomID)
		}
	}
}

// Broadcast sends a message to be processed by the hub.
func (h *Hub) Broadcast(messageType string, payload interface{}, roomID string) {
	h.broadcast <- &Message{
		Type:    messageType,
		Payload: payload,
		RoomID:  roomID,
	}
}
