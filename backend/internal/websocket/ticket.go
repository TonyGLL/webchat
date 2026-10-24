package websocket

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	shared_http "backend/internal/shared/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

const (
	ticketTTL       = 30 * time.Second
	ticketKeyPrefix = "ws_ticket:"
)

// ErrInvalidTicket is returned when a ticket is unknown, expired or already used.
var ErrInvalidTicket = errors.New("invalid or expired websocket ticket")

// TicketStore issues short-lived, single-use tickets that let an authenticated user
// open a websocket without putting their JWT in the URL.
type TicketStore interface {
	Issue(ctx context.Context, userID int) (string, error)
	Redeem(ctx context.Context, ticket string) (int, error)
}

// RedisTicketStore keeps tickets in Redis with a short TTL.
type RedisTicketStore struct {
	client *redis.Client
}

func NewRedisTicketStore(client *redis.Client) *RedisTicketStore {
	return &RedisTicketStore{client: client}
}

func newTicket() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *RedisTicketStore) Issue(ctx context.Context, userID int) (string, error) {
	ticket, err := newTicket()
	if err != nil {
		return "", err
	}
	if err := s.client.Set(ctx, ticketKeyPrefix+ticket, userID, ticketTTL).Err(); err != nil {
		return "", fmt.Errorf("could not store websocket ticket: %w", err)
	}
	return ticket, nil
}

// Redeem consumes the ticket (it can only be used once) and returns its user ID.
func (s *RedisTicketStore) Redeem(ctx context.Context, ticket string) (int, error) {
	value, err := s.client.GetDel(ctx, ticketKeyPrefix+ticket).Result()
	if errors.Is(err, redis.Nil) {
		return 0, ErrInvalidTicket
	}
	if err != nil {
		return 0, err
	}
	userID, err := strconv.Atoi(value)
	if err != nil {
		return 0, ErrInvalidTicket
	}
	return userID, nil
}

// IssueTicket is the handler for POST /ws-ticket. It must run behind the JWT middleware.
func IssueTicket(tickets TicketStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get(shared_http.CtxUserIDKey)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		id, ok := userID.(int)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		ticket, err := tickets.Issue(c.Request.Context(), id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "could not issue ticket"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ticket": ticket})
	}
}
