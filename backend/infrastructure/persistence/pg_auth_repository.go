package persistence

import (
	"backend/application/dtos"
	"backend/domain"
	"context"
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

type PgAuthRepository struct {
	queries *Queries
}

func NewPgAuthRepository(db *sql.DB) *PgAuthRepository {
	return &PgAuthRepository{queries: New(db)}
}

const getUserByEmailQuery = `SELECT id, name, last_name, username, email, phone, avatar_url, last_access, deleted, google_sub, email_verified_at, created_at, updated_at FROM users u WHERE u.email = $1;`

func (r *PgAuthRepository) GetUserByEmail(ctx context.Context, arg *dtos.GetUserByEmailDTO) (*domain.User, error) {
	row := r.queries.db.QueryRowContext(ctx, getUserByEmailQuery, arg.Email)
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

func (r *PgAuthRepository) Register(user *domain.User, password string) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// In a real application, you would insert the user and the hashed password
	// into the database.
	// For now, we'll just return a mock user.
	newUser := &domain.User{
		ID:       2,
		Email:    user.Email,
		Name:     user.Name,
		LastName: user.LastName,
	}

	// This is where you would save the hashedPassword to the Password table
	_ = hashedPassword

	return newUser, nil
}
