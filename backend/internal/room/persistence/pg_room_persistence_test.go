package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"backend/internal/shared/infra/db/dbtest"

	"github.com/google/uuid"
)

func seedRoom(t *testing.T, db *sql.DB, members int) (room uuid.UUID, users []int) {
	t.Helper()
	for i := 0; i < members; i++ {
		var id int
		if err := db.QueryRow(`INSERT INTO users (name, username, email) VALUES ('n', $1, $2) RETURNING id`,
			fmt.Sprintf("u%d", i), fmt.Sprintf("u%d@x.io", i)).Scan(&id); err != nil {
			t.Fatal(err)
		}
		users = append(users, id)
	}
	if err := db.QueryRow(`INSERT INTO rooms (owner_id, name) VALUES ($1, 'r') RETURNING id`, users[0]).Scan(&room); err != nil {
		t.Fatal(err)
	}
	return room, users
}

func TestMarkReadThrottlesWritesButKeepsMembershipAnswer(t *testing.T) {
	db := dbtest.New(t)
	ctx := context.Background()
	room, users := seedRoom(t, db, 2)
	if _, err := db.Exec(`INSERT INTO room_members (room_id, user_id) VALUES ($1, $2)`, room, users[0]); err != nil {
		t.Fatal(err)
	}
	repo := NewPgRoomMemberRepository(db)
	lastRead := func() time.Time {
		var ts time.Time
		if err := db.QueryRow(`SELECT last_read_at FROM room_members WHERE room_id = $1 AND user_id = $2`, room, users[0]).Scan(&ts); err != nil {
			t.Fatal(err)
		}
		return ts
	}

	// stale mark: it is written
	old := time.Now().Add(-time.Minute)
	if _, err := db.Exec(`UPDATE room_members SET last_read_at = $3 WHERE room_id = $1 AND user_id = $2`, room, users[0], old); err != nil {
		t.Fatal(err)
	}
	if ok, err := repo.MarkRead(ctx, room, users[0]); err != nil || !ok {
		t.Fatalf("MarkRead = %v, %v", ok, err)
	}
	first := lastRead()
	if !first.After(old.Add(30 * time.Second)) {
		t.Fatalf("stale last_read_at was not refreshed: %v", first)
	}

	// immediate second mark: no write (same timestamp) and still a member
	if ok, err := repo.MarkRead(ctx, room, users[0]); err != nil || !ok {
		t.Fatalf("throttled MarkRead = %v, %v", ok, err)
	}
	if got := lastRead(); !got.Equal(first) {
		t.Fatalf("last_read_at moved within the 2 s window: %v -> %v", first, got)
	}

	// non-member still reports false
	if ok, err := repo.MarkRead(ctx, room, users[1]); err != nil || ok {
		t.Fatalf("non-member MarkRead = %v, %v; want false", ok, err)
	}
}

func TestListPendingIsCapped(t *testing.T) {
	db := dbtest.New(t)
	ctx := context.Background()
	room, users := seedRoom(t, db, 1)
	invitee := users[0]
	for i := 0; i < maxPendingInvitations+5; i++ {
		var r uuid.UUID
		if err := db.QueryRow(`INSERT INTO rooms (owner_id, name) VALUES ($1, $2) RETURNING id`, users[0], fmt.Sprintf("r%d", i)).Scan(&r); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO room_invitations (room_id, inviter_id, invitee_id) VALUES ($1, $2, $2)`, r, invitee); err != nil {
			t.Fatal(err)
		}
	}
	_ = room
	got, err := NewPgInvitationRepository(db).ListPending(ctx, invitee)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != maxPendingInvitations {
		t.Fatalf("ListPending returned %d, want cap %d", len(got), maxPendingInvitations)
	}
}
