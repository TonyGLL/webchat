package persistence

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"backend/internal/moderation/application"
	"backend/internal/moderation/domain"
	shared_domain "backend/internal/shared/domain"
	"backend/internal/shared/infra/db/dbtest"
	"backend/internal/shared/permissions"

	"github.com/google/uuid"
)

// Against real SQL: the issuer's rank is stored and enforced, and the migration backfill gives
// existing bans their issuer's current rank.
func TestBanRankIsStoredAndEnforced(t *testing.T) {
	db := dbtest.New(t)
	ctx := context.Background()
	for _, u := range []string{"owner", "admin", "mod", "target"} {
		if _, err := db.Exec(`INSERT INTO users (name, username, email) VALUES ($1, $1, $2)`, u, u+"@example.com"); err != nil {
			t.Fatal(err)
		}
	}
	var room uuid.UUID
	if err := db.QueryRow(`INSERT INTO rooms (owner_id, name) VALUES (1, 'r') RETURNING id`).Scan(&room); err != nil {
		t.Fatal(err)
	}
	for id, role := range map[int]string{1: "owner", 2: "admin", 3: "moderator"} {
		if _, err := db.Exec(`INSERT INTO room_members (room_id, user_id, role) VALUES ($1, $2, $3::room_role)`, room, id, role); err != nil {
			t.Fatal(err)
		}
	}
	repo := NewPgRepository(db)
	uc := application.NewBansUseCase(repo, nil)

	if err := uc.Ban(ctx, room, 2, 4, "x", ""); err != nil { // admin bans a stranger
		t.Fatal(err)
	}
	ban, err := repo.FindBan(ctx, room, 4)
	if err != nil || ban == nil || ban.BannerRank != permissions.RankAdmin {
		t.Fatalf("stored ban = %+v, err %v; want banner rank admin", ban, err)
	}
	if err := uc.Unban(ctx, room, 3, 4); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Fatalf("moderator lifting an admin's ban: %v", err)
	}
	if err := uc.Ban(ctx, room, 3, 4, "shorter", "1m"); !errors.Is(err, shared_domain.ErrForbidden) {
		t.Fatalf("moderator overwriting an admin's ban: %v", err)
	}
	if err := uc.Unban(ctx, room, 1, 4); err != nil { // the owner outranks the admin
		t.Fatalf("owner lifting it: %v", err)
	}

	// backfill: a legacy ban (rank 0) from the admin gets the admin's current rank
	if _, err := db.Exec(`INSERT INTO bans (user_id, room_id, banned_by_id, reason) VALUES (4, $1, 2, 'legacy')`, room); err != nil {
		t.Fatal(err)
	}
	migration, err := os.ReadFile("../../../migrations/000017_authz_invitations_bans.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	backfill := string(migration)[strings.Index(string(migration), "UPDATE bans"):]
	if _, err := db.Exec(backfill); err != nil {
		t.Fatalf("backfill statement: %v", err)
	}
	var _ domain.Repository = repo
	legacy, _ := repo.FindBan(ctx, room, 4)
	if legacy == nil || legacy.BannerRank != permissions.RankAdmin {
		t.Fatalf("backfilled ban = %+v", legacy)
	}
}
