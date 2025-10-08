package persistence

import (
	"backend/application/dtos"
	"backend/domain"
	"context"
	"database/sql"
)

type PgAuthRepository struct {
	queries *Queries
}

func NewPgAuthRepository(db *sql.DB) *PgAuthRepository {
	return &PgAuthRepository{queries: New(db)}
}

const getUserByEmailOrUsernameQuery = `SELECT id, name, last_name, username, email, phone, avatar_url, last_access, deleted, google_sub, email_verified_at, created_at, updated_at FROM users u WHERE (u.email = $1 OR u.username = $1) AND u.deleted = FALSE AND u.email_verified_at IS NOT NULL;`

func (r *PgAuthRepository) GetUserByEmailOrUsername(ctx context.Context, arg *dtos.GetUserByEmailOrUsernameDTO) (*domain.User, error) {
	row := r.queries.db.QueryRowContext(ctx, getUserByEmailOrUsernameQuery, arg.User)
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
	return user, err
}

const validatePasswordQuery = `SELECT p.hash FROM passwords p WHERE p.user_id = $1;`

func (r *PgAuthRepository) ValidateUserPasword(ctx context.Context, id int) (string, error) {
	row := r.queries.db.QueryRowContext(ctx, validatePasswordQuery, id)
	var hash string
	err := row.Scan(&hash)
	if err != nil {
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
	row := r.queries.db.QueryRowContext(ctx, registerUserQuery, user.Email, user.Name, user.LastName, user.Username, user.Phone)
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
		return nil, err
	}
	return newUser, nil
}

const createPasswordQuery = `INSERT INTO passwords (user_id, hash) VALUES ($1, $2);`

func (r *PgAuthRepository) CreatePassword(ctx context.Context, userID int, hashedPassword string) error {
	_, err := r.queries.db.ExecContext(ctx, createPasswordQuery, userID, hashedPassword)
	return err
}

const setLastAccessQuery = `UPDATE users SET last_access = NOW() WHERE id = $1;`

func (r *PgAuthRepository) SetLastAccess(ctx context.Context, id int) error {
	_, err := r.queries.db.ExecContext(ctx, setLastAccessQuery, id)
	return err
}
