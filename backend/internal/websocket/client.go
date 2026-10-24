package websocket

import (
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
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	// maxMessageSize caps what a client may send us. Clients send nothing today (messages go
	// through the REST API); raise it if client events such as "typing" or acks are added.
	maxMessageSize = 512
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
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// We are not processing any message from the client for now.
		// The client-to-server communication will happen via REST API.
		// This pump is mainly for handling connection lifecycle (e.g., pong, close).
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
func ServeWs(hub *Hub, upgrader *websocket.Upgrader, tickets TicketStore, roomLister RoomLister, ctx *gin.Context) {
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
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		UserID: id,
		ID:     uuid.New().String(),
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
		log.Printf("Client %s subscribed to %d rooms", client.ID, len(roomIDs))
	}()

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
