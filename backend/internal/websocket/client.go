package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	// maxMessageSize caps what a client may send us. Chat messages go through the REST API;
	// over the socket clients only send small events such as "typing".
	maxMessageSize = 512
	// typingInterval limits how often one client may announce typing in the same room.
	typingInterval = 2 * time.Second
)

// NewUpgrader builds the websocket upgrader. Browser connections are only accepted
// from the allowed origins (the same list used for CORS).
func NewUpgrader(allowedOrigins []string) *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     originChecker(allowedOrigins),
	}
}

// originChecker returns a CheckOrigin function for the given allow-list.
//   - No Origin header (non-browser clients such as curl): allowed; they still need a valid JWT.
//   - Empty allow-list: only same-origin requests are accepted.
func originChecker(allowedOrigins []string) func(r *http.Request) bool {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[strings.TrimRight(o, "/")] = struct{}{}
	}

	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		if len(allowed) == 0 {
			u, err := url.Parse(origin)
			return err == nil && strings.EqualFold(u.Host, r.Host)
		}
		_, ok := allowed[strings.TrimRight(origin, "/")]
		return ok
	}
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub
	// The websocket connection.
	conn *websocket.Conn
	// Buffered channel of outbound messages.
	send chan []byte
	// The user ID of the authenticated user.
	UserID int
	// Unique ID for the client connection.
	ID string
	// presence, when set, is told about this connection's heartbeat.
	presence PresenceStore
	// lastTyping is only touched by readPump.
	lastTyping map[string]time.Time
}

// clientEvent is what a client may send over the socket.
type clientEvent struct {
	Type   string `json:"type"`
	RoomID string `json:"room_id"`
}

// handleClientEvent processes one event from the client. Unknown or malformed events are ignored.
func (c *Client) handleClientEvent(data []byte, now time.Time) {
	var event clientEvent
	if err := json.Unmarshal(data, &event); err != nil || event.Type != "typing" {
		return
	}
	// Only members (subscribed clients) may announce typing, and not more often than typingInterval.
	if !c.hub.IsClientInRoom(c, event.RoomID) {
		return
	}
	if last, ok := c.lastTyping[event.RoomID]; ok && now.Sub(last) < typingInterval {
		return
	}
	if c.lastTyping == nil {
		c.lastTyping = map[string]time.Time{}
	}
	c.lastTyping[event.RoomID] = now
	c.hub.Broadcast("typing", map[string]interface{}{"room_id": event.RoomID, "user_id": c.UserID}, event.RoomID)
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.announceOffline()
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		if c.presence != nil {
			_ = c.presence.Touch(context.Background(), c.UserID, c.ID)
		}
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.handleClientEvent(data, time.Now())
	}
}

// announceOnline marks the connection as present and, if it is the user's first one,
// tells the rooms they belong to.
func (c *Client) announceOnline(roomIDs []string) {
	if c.presence == nil {
		return
	}
	first, err := c.presence.Connect(context.Background(), c.UserID, c.ID)
	if err != nil {
		log.Printf("presence connect failed: %v", err)
		return
	}
	if first {
		c.broadcastPresence(roomIDs, true)
	}
}

// announceOffline removes the connection and, if it was the user's last one, tells their rooms.
func (c *Client) announceOffline() {
	if c.presence == nil {
		return
	}
	rooms := c.hub.RoomsOf(c)
	last, err := c.presence.Disconnect(context.Background(), c.UserID, c.ID)
	if err != nil {
		log.Printf("presence disconnect failed: %v", err)
		return
	}
	if last {
		c.broadcastPresence(rooms, false)
	}
}

func (c *Client) broadcastPresence(roomIDs []string, online bool) {
	for _, roomID := range roomIDs {
		c.hub.Broadcast("presence", map[string]interface{}{"user_id": c.UserID, "online": online}, roomID)
	}
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWs handles websocket requests from the peer. The user authenticates with a
// single-use ticket (see IssueTicket) so the JWT never travels in the URL.
func ServeWs(hub *Hub, upgrader *websocket.Upgrader, tickets TicketStore, roomLister RoomLister, presence PresenceStore, ctx *gin.Context) {
	ticket := ctx.Query("ticket")
	if ticket == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "ticket is required"})
		return
	}

	id, err := tickets.Redeem(ctx.Request.Context(), ticket)
	if err != nil {
		if !errors.Is(err, ErrInvalidTicket) {
			log.Printf("websocket ticket error: %v", err)
		}
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired ticket"})
		return
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		UserID:   id,
		ID:       uuid.New().String(),
		presence: presence,
	}
	hub.Register(client)

	// Subscribe the client to their rooms.
	go func() {
		roomIDs, err := roomLister.GetUserRoomIDs(ctx, id)
		if err != nil {
			log.Printf("failed to get user rooms for client %s: %v", client.ID, err)
			return
		}
		for _, roomID := range roomIDs {
			hub.AddClientToRoom(client, roomID)
		}
		client.announceOnline(roomIDs)
		log.Printf("Client %s subscribed to %d rooms", client.ID, len(roomIDs))
	}()

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
