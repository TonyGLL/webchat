package domain

import "time"

type Password struct {
	UserID    int       `json:"user_id"`
	Hash      string    `json:"hash"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID              int        `json:"id"`
	Name            string     `json:"name"`
	LastName        string     `json:"last_name"`
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	Phone           string     `json:"phone"`
	AvatarUrl       string     `json:"avatar_url"`
	LastAccess      time.Time  `json:"last_access"`
	Deleted         bool       `json:"deleted"`
	GoogleSub       string     `json:"google_sub"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
