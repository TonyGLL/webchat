package persistence

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"backend/internal/room/domain"
	"backend/internal/shared/infra/db/dbtest"

	"github.com/google/uuid"
)

func seedInvitationFixture(t *testing.T, db *sql.DB) (roomID uuid.UUID) {
	t.Helper()
	for i, u := range []string{"owner", "guest", "other"} {
		if _, err := db.Exec(`INSERT INTO users (name, username, email) VALUES ($1, $1, $2)`, u, u+"@example.com"); err != nil {
			t.Fatalf("user %d: %v", i, err)
		}
	}
	if err := db.QueryRow(`INSERT INTO rooms (owner_id, name) VALUES (1, 'r') RETURNING id`).Scan(&roomID); err != nil {
		t.Fatal(err)
	}
	return roomID
}

func TestInvitationsExpireAndCanBeCancelled(t *testing.T) {
	db := dbtest.New(t)
	roomID := seedInvitationFixture(t, db)
	repo := NewPgInvitationRepository(db)
	ctx := context.Background()

	inv, err := repo.Create(ctx, roomID, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !inv.ExpiresAt.After(inv.CreatedAt) {
		t.Fatalf("an invitation needs an expiry, got %v", inv.ExpiresAt)
	}
	if got, _ := repo.FindPending(ctx, inv.ID); got == nil {
		t.Fatal("fresh invitation must be pending")
	}
	if list, _ := repo.ListByRoom(ctx, roomID); len(list) != 1 || list[0].InviteeName != "guest" {
		t.Fatalf("room listing: %+v", list)
	}

	// expire it: it disappears everywhere and no longer blocks a new invitation
	if _, err := db.Exec(`UPDATE room_invitations SET expires_at = NOW() - INTERVAL '1 minute' WHERE id = $1`, inv.ID); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.FindPending(ctx, inv.ID); got != nil {
		t.Fatal("an expired invitation must not be found")
	}
	if list, _ := repo.ListPending(ctx, 2); len(list) != 0 {
		t.Fatalf("expired invitation listed for the invitee: %+v", list)
	}
	if list, _ := repo.ListByRoom(ctx, roomID); len(list) != 0 {
		t.Fatalf("expired invitation listed for the room: %+v", list)
	}
	again, err := repo.Create(ctx, roomID, 1, 2)
	if err != nil {
		t.Fatalf("inviting again after expiry: %v", err)
	}

	// duplicates of a live invitation are still refused
	if _, err := repo.Create(ctx, roomID, 1, 2); !errors.Is(err, domain.ErrAlreadyInvited) {
		t.Fatalf("duplicate: err = %v", err)
	}

	// cancelling is scoped to the room and works once
	if ok, _ := repo.Cancel(ctx, uuid.New(), again.ID); ok {
		t.Fatal("cancel with another room id must not match")
	}
	if ok, err := repo.Cancel(ctx, roomID, again.ID); err != nil || !ok {
		t.Fatalf("cancel: %v %v", ok, err)
	}
	if got, _ := repo.FindPending(ctx, again.ID); got != nil {
		t.Fatal("a cancelled invitation must not be acceptable")
	}
	if ok, _ := repo.Cancel(ctx, roomID, again.ID); ok {
		t.Fatal("cancelling twice must report false")
	}
}
