package dtos

import "time"

// UserResponseDTO represents the user data sent in API responses.
// It omits sensitive information and uses appropriate json tags.
type UserResponseDTO struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	LastName  string    `json:"lastName"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	AvatarURL *string   `json:"avatarUrl,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// AuthResponseDTO is the unified response for both login and registration.
type AuthResponseDTO struct {
	User         UserResponseDTO `json:"user"`
	Token        string          `json:"token"`
	RefreshToken string          `json:"refreshToken"`
}
