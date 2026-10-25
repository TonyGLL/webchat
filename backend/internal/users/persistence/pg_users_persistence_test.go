package persistence

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	shared_domain "backend/internal/shared/domain"
	"backend/internal/shared/infra/db/dbtest"
	users_domain "backend/internal/users/domain"
)

func TestListDirectoryWithAndWithoutQuery(t *testing.T) {
	db := dbtest.New(t)
	ctx := context.Background()
	for _, u := range [][3]string{{"Zoe", "zoe", "zoe@x.io"}, {"Ana", "ana", "ana@x.io"}, {"Bob", "bobby", "bob@x.io"}, {"Cy", "cy", "cy@x.io"}} {
		if _, err := db.Exec(`INSERT INTO users (name, username, email) VALUES ($1, $2, $3)`, u[0], u[1], u[2]); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`UPDATE users SET deleted = TRUE WHERE username = 'cy'`); err != nil {
		t.Fatal(err)
	}
	var self int
	if err := db.QueryRow(`SELECT id FROM users WHERE username = 'zoe'`).Scan(&self); err != nil {
		t.Fatal(err)
	}
	repo := NewPgUsersRepository(db)

	names := func(es []*users_domain.DirectoryEntry) (out []string) {
		for _, e := range es {
			out = append(out, e.Username)
		}
		return
	}
	eq := func(got []string, want ...string) {
		t.Helper()
		if len(got) != len(want) {
			t.Fatalf("got %v, want %v", got, want)
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("got %v, want %v", got, want)
			}
		}
	}

	all, err := repo.ListDirectory(ctx, self, "", 50)
	if err != nil {
		t.Fatal(err)
	}
	eq(names(all), "ana", "bobby") // ordered by name; excludes self and deactivated

	ws, err := repo.ListDirectory(ctx, self, "   ", 1)
	if err != nil {
		t.Fatal(err)
	}
	eq(names(ws), "ana")

	hit, err := repo.ListDirectory(ctx, self, "OB", 50)
	if err != nil {
		t.Fatal(err)
	}
	eq(names(hit), "bobby")

	var ana int
	_ = db.QueryRow(`SELECT id FROM users WHERE username = 'ana'`).Scan(&ana)
	var bob int
	_ = db.QueryRow(`SELECT id FROM users WHERE username = 'bobby'`).Scan(&bob)
	byIDs, err := repo.ListDirectoryByIDs(ctx, self, "", []int{bob, ana, self}, 50)
	if err != nil {
		t.Fatal(err)
	}
	eq(names(byIDs), "ana", "bobby")
	byIDs, err = repo.ListDirectoryByIDs(ctx, self, "bo", []int{bob, ana}, 50)
	if err != nil {
		t.Fatal(err)
	}
	eq(names(byIDs), "bobby")
}

func TestExportAccountStreamsValidJSON(t *testing.T) {
	db := dbtest.New(t)
	ctx := context.Background()
	var uid int
	if err := db.QueryRow(`INSERT INTO users (name, username, email) VALUES ('Ana', 'ana', 'ana@x.io') RETURNING id`).Scan(&uid); err != nil {
		t.Fatal(err)
	}
	var room string
	if err := db.QueryRow(`INSERT INTO rooms (owner_id, name) VALUES ($1, 'r <&> "q"') RETURNING id`, uid).Scan(&room); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO room_members (room_id, user_id, role) VALUES ($1, $2, 'owner')`, room, uid); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		if _, err := db.Exec(`INSERT INTO messages (content, author_id, room_id) VALUES ($1, $2, $3)`, "hola \"mundo\" <b>", uid, room); err != nil {
			t.Fatal(err)
		}
	}

	var buf bytes.Buffer
	if err := NewPgUsersRepository(db).ExportAccount(ctx, uid, &buf); err != nil {
		t.Fatal(err)
	}
	var got users_domain.AccountExport
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("export is not valid JSON: %v\n%s", err, buf.String())
	}
	if got.Profile.Username != "ana" || len(got.Rooms) != 1 || len(got.Messages) != 5 {
		t.Fatalf("unexpected export: %+v", got)
	}
	if got.Rooms[0].Name != `r <&> "q"` || got.Messages[0].Content != `hola "mundo" <b>` {
		t.Fatalf("content not preserved: %+v %+v", got.Rooms[0], got.Messages[0])
	}
	if got.Reactions == nil || got.Reports == nil {
		t.Fatal("empty sections must be arrays, not null")
	}

	buf.Reset()
	err := NewPgUsersRepository(db).ExportAccount(ctx, uid+999, &buf)
	if !errors.Is(err, shared_domain.ErrNotFound) || buf.Len() != 0 {
		t.Fatalf("unknown user: err=%v, wrote %d bytes; want ErrNotFound and nothing written", err, buf.Len())
	}
}
