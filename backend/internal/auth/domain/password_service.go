package domain

import "golang.org/x/crypto/bcrypt"

// PasswordService defines the interface for password operations.
type PasswordService interface {
	// HashPassword generates a bcrypt hash from a password.
	HashPassword(password string) (string, error)
	// CheckPasswordHash compares a password with a hash.
	CheckPasswordHash(password, hash string) bool
}

// bcryptPasswordService is an implementation of PasswordService using bcrypt.
type bcryptPasswordService struct{}

// NewPasswordService creates a new instance of bcryptPasswordService.
func NewPasswordService() PasswordService {
	return &bcryptPasswordService{}
}

func (s *bcryptPasswordService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *bcryptPasswordService) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}