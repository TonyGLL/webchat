package websocket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type fakeTickets struct{ user int }

func (f fakeTickets) Issue(context.Context, int) (string, error)  { return "t", nil }
func (f fakeTickets) Redeem(context.Context, string) (int, error) { return f.user, nil }

type fakeGuard struct {
	valid bool
	asked []int
}

func (g *fakeGuard) Valid(_ context.Context, id int, _ time.Time) bool {
	g.asked = append(g.asked, id)
	return g.valid
}

func TestServeWsRefusesUsersTheGuardRejects(t *testing.T) {
	gin.SetMode(gin.TestMode)
	guard := &fakeGuard{valid: false}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/ws?ticket=t", nil)

	ServeWs(NewHub(), NewUpgrader(nil), fakeTickets{user: 9}, nil, nil, guard, c)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if len(guard.asked) != 1 || guard.asked[0] != 9 {
		t.Fatalf("guard must be asked about the ticket's user, got %v", guard.asked)
	}
}
