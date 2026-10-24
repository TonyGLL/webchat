package persistence

import (
	"context"
	"strings"

	"backend/internal/users/domain"
)

// likeEscaper neutralizes LIKE wildcards in user input.
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func (r *PgUsersRepository) ListDirectory(ctx context.Context, selfID int, q string, limit int) ([]*domain.DirectoryEntry, error) {
	pattern := "%" + likeEscaper.Replace(strings.TrimSpace(q)) + "%"
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, username, name, last_name, avatar_url
		FROM users
		WHERE deleted = FALSE AND id <> $1
		  AND (username ILIKE $2 OR name ILIKE $2 OR COALESCE(last_name, '') ILIKE $2)
		ORDER BY name, username
		LIMIT $3`, selfID, pattern, limit)
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
