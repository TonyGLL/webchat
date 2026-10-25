package persistence

import (
	"context"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"backend/internal/users/domain"
)

// likeEscaper neutralizes LIKE wildcards in user input.
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// The filter expression must match idx_users_directory_trgm (migration 000014) exactly,
// otherwise the planner cannot use the trigram index.
const directorySelect = `
	SELECT id, username, name, last_name, avatar_url
	FROM users
	WHERE deleted = FALSE AND id <> $1`

const directoryLike = `
	  AND lower(username || ' ' || name || ' ' || COALESCE(last_name, '')) LIKE lower($2)`

const directoryOrder = ` ORDER BY name, username`

func directoryPattern(q string) string {
	return "%" + likeEscaper.Replace(strings.TrimSpace(q)) + "%"
}

// directoryFilter returns the LIKE clause and its argument; an empty q matches everybody, so the
// clause is dropped and the ordered partial index (migration 000016) serves the default listing.
func directoryFilter(q string) (clause string, args []any) {
	if strings.TrimSpace(q) == "" {
		return "", nil
	}
	return directoryLike, []any{directoryPattern(q)}
}

func (r *PgUsersRepository) ListDirectory(ctx context.Context, selfID int, q string, limit int) ([]*domain.DirectoryEntry, error) {
	clause, args := directoryFilter(q)
	args = append([]any{selfID}, append(args, limit)...)
	return r.queryDirectory(ctx, directorySelect+clause+directoryOrder+fmt.Sprintf(` LIMIT $%d`, len(args)), args...)
}

func (r *PgUsersRepository) ListDirectoryByIDs(ctx context.Context, selfID int, q string, ids []int, limit int) ([]*domain.DirectoryEntry, error) {
	if len(ids) == 0 {
		return []*domain.DirectoryEntry{}, nil
	}
	arr := make([]int64, len(ids))
	for i, id := range ids {
		arr[i] = int64(id)
	}
	clause, args := directoryFilter(q)
	args = append([]any{selfID}, append(args, pq.Array(arr), limit)...)
	n := len(args)
	return r.queryDirectory(ctx, directorySelect+clause+fmt.Sprintf(` AND id = ANY($%d)`, n-1)+directoryOrder+fmt.Sprintf(` LIMIT $%d`, n), args...)
}

func (r *PgUsersRepository) queryDirectory(ctx context.Context, query string, args ...any) ([]*domain.DirectoryEntry, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	entries := []*domain.DirectoryEntry{}
	for rows.Next() {
		e := &domain.DirectoryEntry{}
		if err := rows.Scan(&e.ID, &e.Username, &e.Name, &e.LastName, &e.AvatarURL); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
