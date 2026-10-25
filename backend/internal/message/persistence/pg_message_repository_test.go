package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"backend/internal/message/domain"
	"backend/internal/shared/infra/db/dbtest"

	"github.com/google/uuid"
)

// legacyHistoryQuery is the pre-BP-1 query (aggregate below the LIMIT), kept as the oracle.
const legacyHistoryQuery = `
	SELECT m.id, m.content, COALESCE(m.author_id, 0), m.room_id, m.reply_to_message_id, m.created_at,
	       m.edited_at, m.deleted_at, m.pinned_at,
	       COALESCE(json_agg(json_build_object('message_id', r.message_id, 'user_id', r.user_id,
	                'emoji', r.emoji, 'created_at', r.created_at)) FILTER (WHERE r.message_id IS NOT NULL), '[]') AS reactions
	FROM messages m
	LEFT JOIN reactions r ON m.id = r.message_id
	WHERE m.room_id = $1
	  AND m.deleted_at IS NULL
	  AND ($2::uuid IS NULL OR (m.created_at, m.id) < (SELECT created_at, id FROM messages WHERE id = $2))
	GROUP BY m.id
	ORDER BY m.created_at DESC, m.id DESC
	LIMIT $3`

func legacyFind(t *testing.T, db *sql.DB, room uuid.UUID, limit int, before *uuid.UUID) []*domain.Message {
	t.Helper()
	rows, err := db.Query(legacyHistoryQuery, room, before, limit)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	out := []*domain.Message{}
	for rows.Next() {
		var m domain.Message
		var raw []byte
		if err := rows.Scan(&m.ID, &m.Content, &m.AuthorID, &m.RoomID, &m.ReplyToMessageID, &m.CreatedAt, &m.EditedAt, &m.DeletedAt, &m.PinnedAt, &raw); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(raw, &m.Reactions); err != nil {
			t.Fatal(err)
		}
		out = append(out, &m)
	}
	return out
}

func TestFindByRoomIDMatchesLegacyQuery(t *testing.T) {
	db := dbtest.New(t)
	ctx := context.Background()

	var u1, u2 int
	for i, dst := range []*int{&u1, &u2} {
		if err := db.QueryRow(`INSERT INTO users (name, username, email) VALUES ('n', $1, $2) RETURNING id`,
			[]string{"a", "b"}[i], []string{"a@x.io", "b@x.io"}[i]).Scan(dst); err != nil {
			t.Fatal(err)
		}
	}
	var room, other uuid.UUID
	for _, dst := range []*uuid.UUID{&room, &other} {
		if err := db.QueryRow(`INSERT INTO rooms (owner_id, name) VALUES ($1, 'r') RETURNING id`, u1).Scan(dst); err != nil {
			t.Fatal(err)
		}
	}

	base := time.Now().Add(-time.Hour).Truncate(time.Second)
	ids := []uuid.UUID{}
	for i := 0; i < 30; i++ {
		// pairs of messages share created_at to exercise the (created_at, id) tie-break
		at := base.Add(time.Duration(i/2) * time.Minute)
		id := uuid.New()
		if _, err := db.Exec(`INSERT INTO messages (id, content, author_id, room_id, created_at) VALUES ($1, $2, $3, $4, $5)`,
			id, "msg", u1, room, at); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	// noise in another room
	if _, err := db.Exec(`INSERT INTO messages (content, author_id, room_id) VALUES ('x', $1, $2)`, u1, other); err != nil {
		t.Fatal(err)
	}
	// deleted message, reactions on several messages (including one with two users)
	if _, err := db.Exec(`UPDATE messages SET deleted_at = NOW() WHERE id = $1`, ids[29]); err != nil {
		t.Fatal(err)
	}
	for _, r := range []struct {
		msg   uuid.UUID
		user  int
		emoji string
	}{{ids[28], u1, "👍"}, {ids[28], u2, "👍"}, {ids[28], u2, "🎉"}, {ids[10], u1, "👍"}, {ids[29], u1, "👍"}} {
		if _, err := db.Exec(`INSERT INTO reactions (message_id, user_id, emoji) VALUES ($1, $2, $3)`, r.msg, r.user, r.emoji); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`UPDATE messages SET pinned_at = NOW(), edited_at = NOW() WHERE id = $1`, ids[27]); err != nil {
		t.Fatal(err)
	}

	repo := NewPgMessageRepository(db)
	compare := func(name string, limit int, before *uuid.UUID) []*domain.Message {
		t.Helper()
		got, err := repo.FindByRoomID(ctx, room, limit, before)
		if err != nil {
			t.Fatal(err)
		}
		want := legacyFind(t, db, room, limit, before)
		if len(got) != len(want) {
			t.Fatalf("%s: %d rows, legacy %d", name, len(got), len(want))
		}
		for i := range got {
			g, _ := json.Marshal(got[i])
			w, _ := json.Marshal(want[i])
			if got[i].ID != want[i].ID {
				t.Fatalf("%s: row %d is %s, legacy %s", name, i, got[i].ID, want[i].ID)
			}
			// reactions come from an unordered aggregate: compare as sets
			if len(got[i].Reactions) != len(want[i].Reactions) {
				t.Fatalf("%s: reactions differ on %s: %s vs %s", name, got[i].ID, g, w)
			}
			set := map[string]bool{}
			for _, r := range want[i].Reactions {
				set[r.Emoji+string(rune(r.UserID))] = true
			}
			for _, r := range got[i].Reactions {
				if !set[r.Emoji+string(rune(r.UserID))] {
					t.Fatalf("%s: unexpected reaction %+v on %s", name, r, got[i].ID)
				}
			}
		}
		return got
	}

	first := compare("page 1", 10, nil)
	if len(first) != 10 {
		t.Fatalf("page 1 has %d rows", len(first))
	}
	for _, m := range first {
		if m.ID == ids[29] {
			t.Fatal("deleted message returned")
		}
	}
	if len(first[0].Reactions) != 3 { // ids[28] is the newest visible message
		t.Fatalf("newest message has %d reactions, want 3", len(first[0].Reactions))
	}
	// walk every page with the before cursor, including odd limits and a tie boundary
	for _, limit := range []int{1, 3, 7, 50} {
		var cursor *uuid.UUID
		total := 0
		for {
			page := compare("cursor", limit, cursor)
			total += len(page)
			if len(page) < limit {
				break
			}
			c := page[len(page)-1].ID
			cursor = &c
		}
		if total != 29 {
			t.Fatalf("limit %d walked %d messages, want 29", limit, total)
		}
	}
}
