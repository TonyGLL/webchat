// Package websocket handles real-time communication using WebSockets.
package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients in specific rooms.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan *Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// A map where keys are room IDs and values are maps of clients in that room.
	rooms map[string]map[*Client]bool

	// Mutex to protect concurrent access to the rooms map.
	mu sync.Mutex
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
	}
}

// Run starts the hub's event loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.removeClientFromAllRooms(client)
			}
		case message := <-h.broadcast:
			h.handleBroadcast(message)
		}
	}
}

// handleBroadcast sends a message to all clients in the specified room.
func (h *Hub) handleBroadcast(message *Message) {
	h.mu.Lock()
	defer h.mu.Unlock()

	roomID := message.RoomID
	if room, ok := h.rooms[roomID]; ok {
		// Convert the message payload to JSON
		payloadBytes, err := json.Marshal(message.Payload)
		if err != nil {
			log.Printf("error marshalling broadcast message payload: %v", err)
			return
		}

		// Create the final message structure
		finalMessage := WebsocketMessage{
			Type:    message.Type,
			Payload: json.RawMessage(payloadBytes),
		}

		// Serialize the final message
		finalBytes, err := json.Marshal(finalMessage)
		if err != nil {
			log.Printf("error marshalling final websocket message: %v", err)
			return
		}

		for client := range room {
			select {
			case client.send <- finalBytes:
			default:
				close(client.send)
				delete(h.clients, client)
				h.removeClientFromAllRooms(client)
			}
		}
	}
}

// AddClientToRoom adds a client to a specific room.
func (h *Hub) AddClientToRoom(client *Client, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[roomID]; !ok {
		h.rooms[roomID] = make(map[*Client]bool)
	}
	h.rooms[roomID][client] = true
}

// RemoveClientFromRoom removes a client from a specific room.
func (h *Hub) removeClientFromAllRooms(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for roomID := range h.rooms {
		if _, ok := h.rooms[roomID][client]; ok {
			delete(h.rooms[roomID], client)
			if len(h.rooms[roomID]) == 0 {
				delete(h.rooms, roomID)
			}
		}
	}
}

// Broadcast sends a message to be processed by the hub.
func (h *Hub) Broadcast(messageType string, payload interface{}, roomID string) {
	message := &Message{
		Type:    messageType,
		Payload: payload,
		RoomID:  roomID,
	}
	h.broadcast <- message
}
