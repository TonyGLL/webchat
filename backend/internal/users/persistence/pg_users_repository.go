package persistence

import (
	"context"
	"database/sql"
	"errors"

	"backend/internal/auth/domain"
	shared_domain "backend/internal/shared/domain"
	"backend/internal/shared/infra/db"
)

// PgUsersRepository implements the repositories.AuthRepository interface for PostgreSQL.
type PgUsersRepository struct {
	db db.DBTX
}

// NewPgUsersRepository creates a new PgUsersRepository.
// It accepts a DBTX interface, which can be either a *sql.DB or *sql.Tx.
func NewPgUsersRepository(dbtx db.DBTX) *PgUsersRepository {
	return &PgUsersRepository{db: dbtx}
}

const getUserByIDQuery = `SELECT id, name, last_name, username, email, phone, avatar_url, last_access, google_sub, email_verified_at, created_at, updated_at FROM users WHERE id = $1 AND deleted = FALSE;`

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
