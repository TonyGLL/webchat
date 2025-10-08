package persistence

import (
	"context"
	"database/sql"
	"errors"

	"backend/internal/auth/domain"
	shared_domain "backend/internal/shared/domain"
	"backend/internal/shared/infra/db"

	"github.com/lib/pq"
)

// PgAuthRepository implements the repositories.AuthRepository interface for PostgreSQL.
type PgAuthRepository struct {
	db db.DBTX
}

// NewPgAuthRepository creates a new PgAuthRepository.
// It accepts a DBTX interface, which can be either a *sql.DB or *sql.Tx.
func NewPgAuthRepository(dbtx db.DBTX) *PgAuthRepository {
	return &PgAuthRepository{db: dbtx}
}

const getUserByEmailOrUsernameQuery = `SELECT id, name, last_name, username, email, phone, avatar_url, last_access, deleted, google_sub, email_verified_at, created_at, updated_at FROM users u WHERE (u.email = $1 OR u.username = $1) AND u.deleted = FALSE AND u.email_verified_at IS NOT NULL;`

func (r *PgAuthRepository) GetUserByEmailOrUsername(ctx context.Context, emailOrUsername string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, getUserByEmailOrUsernameQuery, emailOrUsername)
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
		&user.Deleted,
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

const validatePasswordQuery = `SELECT p.hash FROM passwords p WHERE p.user_id = $1;`

func (r *PgAuthRepository) ValidateUserPassword(ctx context.Context, id int) (string, error) {
	row := r.db.QueryRowContext(ctx, validatePasswordQuery, id)
	var hash string
	err := row.Scan(&hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", shared_domain.ErrNotFound
		}
		return "", err
	}
	return hash, nil
}

const registerUserQuery = `
	INSERT INTO users (email, name, last_name, username, phone)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, name, last_name, username, email, phone, avatar_url, last_access, deleted, google_sub, email_verified_at, created_at, updated_at;
`

func (r *PgAuthRepository) Register(ctx context.Context, user *domain.User) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, registerUserQuery, user.Email, user.Name, user.LastName, user.Username, user.Phone)
	newUser := &domain.User{}
	err := row.Scan(
		&newUser.ID,
		&newUser.Name,
		&newUser.LastName,
		&newUser.Username,
		&newUser.Email,
		&newUser.Phone,
		&newUser.AvatarUrl,
		&newUser.LastAccess,
		&newUser.Deleted,
		&newUser.GoogleSub,
		&newUser.EmailVerifiedAt,
		&newUser.CreatedAt,
		&newUser.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code.Name() == "unique_violation" {
				return nil, shared_domain.ErrConflict
			}
		}
		return nil, err
	}
	return newUser, nil
}

const createPasswordQuery = `INSERT INTO passwords (user_id, hash) VALUES ($1, $2);`

func (r *PgAuthRepository) CreatePassword(ctx context.Context, userID int, hashedPassword string) error {
	_, err := r.db.ExecContext(ctx, createPasswordQuery, userID, hashedPassword)
	return err
}

const setLastAccessQuery = `UPDATE users SET last_access = NOW() WHERE id = $1;`

func (r *PgAuthRepository) SetLastAccess(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, setLastAccessQuery, id)
	return err
}