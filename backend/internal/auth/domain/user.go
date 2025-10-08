package domain

import "time"

type Password struct {
	UserID    int
	Hash      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type User struct {
	ID              int
	Name            string
	LastName        string
	Username        string
	Email           string
	Phone           string
	AvatarUrl       *string
	LastAccess      time.Time
	Deleted         bool
	GoogleSub       *string
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
