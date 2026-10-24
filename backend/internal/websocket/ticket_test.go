package websocket

import "testing"

func TestNewTicketIsRandomAndLongEnough(t *testing.T) {
	a, err := newTicket()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := newTicket()
	if len(a) != 64 {
		t.Fatalf("ticket length = %d, want 64 hex chars", len(a))
	}
	if a == b {
		t.Fatal("two tickets must not be equal")
	}
}
