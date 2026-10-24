package domain

import "golang.org/x/crypto/bcrypt"

// PasswordService defines the interface for password operations.
type PasswordService interface {
	// HashPassword generates a bcrypt hash from a password.
	HashPassword(password string) (string, error)
	// CheckPasswordHash compares a password with a hash.
	CheckPasswordHash(password, hash string) bool
}

// bcryptCost is the work factor for new hashes. Older hashes (default cost 10) still verify and
// keep their own cost until the user changes the password.
const bcryptCost = 12

// bcryptPasswordService is an implementation of PasswordService using bcrypt.
type bcryptPasswordService struct{}

// NewPasswordService creates a new instance of bcryptPasswordService.
func NewPasswordService() PasswordService {
	return &bcryptPasswordService{}
}

func (s *bcryptPasswordService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *bcryptPasswordService) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
