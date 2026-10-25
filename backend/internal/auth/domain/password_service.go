package domain

import (
	"errors"
	"runtime"
	"time"

	"golang.org/x/crypto/bcrypt"
)

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

// ErrPasswordBusy is returned when every bcrypt slot stays taken for longer than the queue timeout.
var ErrPasswordBusy = errors.New("password hashing is busy, try again shortly")

// hashQueueTimeout is how long a request waits for a free bcrypt slot.
const hashQueueTimeout = 5 * time.Second

// bcryptPasswordService is an implementation of PasswordService using bcrypt. A semaphore bounds
// how many hashes run at once (about GOMAXPROCS/2), so a burst of logins cannot starve the rest
// of the API; excess requests queue and give up after hashQueueTimeout.
type bcryptPasswordService struct {
	slots   chan struct{}
	timeout time.Duration
}

// NewPasswordService creates a new instance of bcryptPasswordService.
func NewPasswordService() PasswordService {
	return newBcryptPasswordService(max(1, runtime.GOMAXPROCS(0)/2), hashQueueTimeout)
}

func newBcryptPasswordService(concurrency int, timeout time.Duration) *bcryptPasswordService {
	return &bcryptPasswordService{slots: make(chan struct{}, concurrency), timeout: timeout}
}

// acquire waits for a free slot; false if the queue timeout expires first.
func (s *bcryptPasswordService) acquire() bool {
	select {
	case s.slots <- struct{}{}:
		return true
	default:
	}
	timer := time.NewTimer(s.timeout)
	defer timer.Stop()
	select {
	case s.slots <- struct{}{}:
		return true
	case <-timer.C:
		return false
	}
}

func (s *bcryptPasswordService) release() { <-s.slots }

func (s *bcryptPasswordService) HashPassword(password string) (string, error) {
	if !s.acquire() {
		return "", ErrPasswordBusy
	}
	defer s.release()
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash reports false when the password does not match or when no slot became free
// in time (the caller treats both as a failed attempt; the client may simply retry).
func (s *bcryptPasswordService) CheckPasswordHash(password, hash string) bool {
	if !s.acquire() {
		return false
	}
	defer s.release()
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
