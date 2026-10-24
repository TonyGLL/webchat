package dtos

import "time"

type UserProfileDTO struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	LastName  string  `json:"last_name"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	AvatarUrl *string `json:"avatar_url"`
	// Role is the global role: user, moderator or admin.
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
