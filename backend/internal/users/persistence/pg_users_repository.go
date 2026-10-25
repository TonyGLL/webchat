package persistence

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"backend/internal/auth/domain"
	shared_domain "backend/internal/shared/domain"
	"backend/internal/shared/infra/db"
	users_domain "backend/internal/users/domain"
)

// PgUsersRepository implements the repositories.AuthRepository interface for PostgreSQL.
type PgUsersRepository struct {
	db db.DBTX
	// sqlDB is only needed for operations that run their own transaction.
	sqlDB *sql.DB
}

// NewPgUsersRepository creates a new PgUsersRepository.
// It accepts a DBTX interface, which can be either a *sql.DB or *sql.Tx.
func NewPgUsersRepository(dbtx db.DBTX) *PgUsersRepository {
	repo := &PgUsersRepository{db: dbtx}
	repo.sqlDB, _ = dbtx.(*sql.DB)
	return repo
}

const getUserByIDQuery = `SELECT id, name, last_name, username, email, phone, avatar_url, last_access, google_sub, email_verified_at, created_at, updated_at, role FROM users WHERE id = $1 AND deleted = FALSE;`

func (r *PgUsersRepository) GetUserByID(ctx context.Context, id int) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, getUserByIDQuery, id)
	user := &domain.User{}
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.LastName,
		&user.Username,
		&user.Email,
		&user.Phone,
		&user.AvatarUrl,
		&user.LastAccess,
		&user.GoogleSub,
		&user.EmailVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Role,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, shared_domain.ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

const deactivateUserQuery = `UPDATE users SET deleted = TRUE, updated_at = NOW() WHERE id = $1 AND deleted = FALSE RETURNING id;`

func (r *PgUsersRepository) DeactivateUser(ctx context.Context, id int) error {
	row := r.db.QueryRowContext(ctx, deactivateUserQuery, id)
	user := &domain.User{}
	err := row.Scan(
		&user.ID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return shared_domain.ErrNotFound
		}
		return err
	}
	return nil
}

func (r *PgUsersRepository) PasswordHash(ctx context.Context, id int) (string, error) {
	var hash string
	err := r.db.QueryRowContext(ctx, `SELECT hash FROM passwords WHERE user_id = $1`, id).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", shared_domain.ErrNotFound
	}
	return hash, err
}

// ExportAccount streams the user's data as one JSON document (same shape as
// domain.AccountExport) straight to w, one row at a time, so memory stays flat however many
// messages the user wrote (BP-13). ErrNotFound is returned before anything is written; an
// error after that leaves a truncated, invalid JSON document and the caller can only log it.
func (r *PgUsersRepository) ExportAccount(ctx context.Context, id int, w io.Writer) error {
	var p users_domain.ExportedProfile
	err := r.db.QueryRowContext(ctx, `SELECT id, name, last_name, username, email, phone, avatar_url, role, created_at, last_access FROM users WHERE id = $1 AND deleted = FALSE`, id).
		Scan(&p.ID, &p.Name, &p.LastName, &p.Username, &p.Email, &p.Phone, &p.AvatarURL, &p.Role, &p.CreatedAt, &p.LastAccess)
	if errors.Is(err, sql.ErrNoRows) {
		return shared_domain.ErrNotFound
	}
	if err != nil {
		return err
	}

	bw := bufio.NewWriterSize(w, 32<<10)
	enc := json.NewEncoder(bw)
	enc.SetEscapeHTML(false)

	exportedAt, _ := json.Marshal(time.Now().UTC())
	if _, err := fmt.Fprintf(bw, `{"exported_at":%s,"profile":`, exportedAt); err != nil {
		return err
	}
	if err := enc.Encode(p); err != nil {
		return err
	}

	// stream writes one JSON array: rows are scanned and encoded one by one.
	stream := func(key, query string, scan func(*sql.Rows) (any, error)) error {
		if _, err := fmt.Fprintf(bw, `,%q:[`, key); err != nil {
			return err
		}
		rows, err := r.db.QueryContext(ctx, query, id)
		if err != nil {
			return err
		}
		defer rows.Close()
		for first := true; rows.Next(); first = false {
			if !first {
				if _, err := bw.WriteString(","); err != nil {
					return err
				}
			}
			v, err := scan(rows)
			if err != nil {
				return err
			}
			if err := enc.Encode(v); err != nil {
				return err
			}
		}
		if err := rows.Err(); err != nil {
			return err
		}
		_, err = bw.WriteString("]")
		return err
	}

	if err := stream("rooms", `
		SELECT r.id, CASE WHEN r.is_direct THEN 'Direct message' ELSE r.name END, r.is_direct, rm.role::text, rm.joined_at
		FROM room_members rm JOIN rooms r ON r.id = rm.room_id WHERE rm.user_id = $1 ORDER BY rm.joined_at`,
		func(rows *sql.Rows) (any, error) {
			var x users_domain.ExportedRoom
			err := rows.Scan(&x.RoomID, &x.Name, &x.IsDirect, &x.Role, &x.JoinedAt)
			return x, err
		}); err != nil {
		return err
	}
	if err := stream("messages", `SELECT id, room_id, content, created_at, edited_at, deleted_at FROM messages WHERE author_id = $1 ORDER BY created_at`,
		func(rows *sql.Rows) (any, error) {
			var x users_domain.ExportedMessage
			err := rows.Scan(&x.ID, &x.RoomID, &x.Content, &x.CreatedAt, &x.EditedAt, &x.DeletedAt)
			return x, err
		}); err != nil {
		return err
	}
	if err := stream("reactions", `SELECT message_id, emoji, created_at FROM reactions WHERE user_id = $1 ORDER BY created_at`,
		func(rows *sql.Rows) (any, error) {
			var x users_domain.ExportedReaction
			err := rows.Scan(&x.MessageID, &x.Emoji, &x.CreatedAt)
			return x, err
		}); err != nil {
		return err
	}
	if err := stream("reports_filed", `SELECT reported_message_id, reason, status::text, created_at FROM reports WHERE reported_by_id = $1 ORDER BY created_at`,
		func(rows *sql.Rows) (any, error) {
			var x users_domain.ExportedReport
			err := rows.Scan(&x.MessageID, &x.Reason, &x.Status, &x.CreatedAt)
			return x, err
		}); err != nil {
		return err
	}
	if _, err := bw.WriteString("}\n"); err != nil {
		return err
	}
	return bw.Flush()
}

func closeRows(rows *sql.Rows) error {
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	return rows.Close()
}

// successorQuery picks who inherits a room: the longest-standing admin, else moderator, else member.
const successorQuery = `
	SELECT user_id FROM room_members
	WHERE room_id = $1 AND user_id <> $2
	ORDER BY CASE role::text WHEN 'admin' THEN 0 WHEN 'moderator' THEN 1 ELSE 2 END, joined_at
	LIMIT 1`

func (r *PgUsersRepository) DeleteAccount(ctx context.Context, id int) error {
	if r.sqlDB == nil {
		return errors.New("account deletion needs a transactional database handle")
	}
	tx, err := r.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after a successful commit

	rows, err := tx.QueryContext(ctx, `SELECT id, is_direct FROM rooms WHERE owner_id = $1`, id)
	if err != nil {
		return err
	}
	type owned struct {
		id       string
		isDirect bool
	}
	var rooms []owned
	for rows.Next() {
		var o owned
		if err := rows.Scan(&o.id, &o.isDirect); err != nil {
			rows.Close()
			return err
		}
		rooms = append(rooms, o)
	}
	if err := closeRows(rows); err != nil {
		return err
	}

	for _, room := range rooms {
		var successor int
		err := tx.QueryRowContext(ctx, successorQuery, room.id, id).Scan(&successor)
		if errors.Is(err, sql.ErrNoRows) {
			if _, err := tx.ExecContext(ctx, `DELETE FROM rooms WHERE id = $1`, room.id); err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE rooms SET owner_id = $2 WHERE id = $1`, room.id, successor); err != nil {
			return err
		}
		if !room.isDirect { // direct conversations have no owner role
			if _, err := tx.ExecContext(ctx, `UPDATE room_members SET role = 'owner' WHERE room_id = $1 AND user_id = $2`, room.id, successor); err != nil {
				return err
			}
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id); err != nil {
		return err
	}
	return tx.Commit()
}
