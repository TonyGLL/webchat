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

const getUserByIDQuery = `SELECT id, name, last_name, username, email, phone, avatar_url, last_access, deleted, google_sub, email_verified_at, created_at, updated_at FROM users u WHERE u.id = $1 AND u.deleted = FALSE;`

func (r *PgAuthRepository) GetUserByID(ctx context.Context, id int) (*domain.User, error) {
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

const getVerifyEmailQuery = `SELECT u.id, u.email FROM users u WHERE u.email = $1 AND u.id = $2 AND u.deleted = FALSE AND u.email_verified_at IS NULL;`

func (r *PgAuthRepository) GetUserVerifyEmail(ctx context.Context, email string, userID int) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, getVerifyEmailQuery, email, userID)
	user := &domain.User{}
	err := row.Scan(
		&user.ID,
		&user.Email,
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
	INSERT INTO users (email, name, last_name, username, phone, last_access)
	VALUES ($1, $2, $3, $4, $5, NOW())
	RETURNING id, name, last_name, username, email, phone, avatar_url, last_access, deleted, google_sub, email_verified_at, created_at, updated_at;
`

const createPasswordQuery = `INSERT INTO passwords (user_id, hash) VALUES ($1, $2);`

// RegisterWithPassword inserts the user and their password in a single transaction,
// so a failure in either step never leaves a user without a password behind.
func (r *PgAuthRepository) RegisterWithPassword(ctx context.Context, user *domain.User, hashedPassword string) (*domain.User, error) {
	pool, ok := r.db.(*sql.DB)
	if !ok {
		return nil, errors.New("register requires a *sql.DB to start a transaction")
	}

	tx, err := pool.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // no-op once committed

	newUser := &domain.User{}
	err = tx.QueryRowContext(ctx, registerUserQuery, user.Email, user.Name, user.LastName, user.Username, user.Phone).Scan(
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
		if errors.As(err, &pqErr) && pqErr.Code.Name() == "unique_violation" {
			return nil, shared_domain.ErrConflict
		}
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, createPasswordQuery, newUser.ID, hashedPassword); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return newUser, nil
}

const setLastAccessQuery = `UPDATE users SET last_access = NOW() WHERE id = $1;`

func (r *PgAuthRepository) SetLastAccess(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, setLastAccessQuery, id)
	return err
}

const verifyEmailQuery = `UPDATE users SET email_verified_at = NOW() WHERE id = $1;`

func (r *PgAuthRepository) VerifyEmail(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, verifyEmailQuery, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return shared_domain.ErrNotFound
	}
	return nil
}
