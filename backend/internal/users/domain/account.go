package domain

import "time"

// AccountExport is everything the application stores about one user (the "download my data" file).
// Password hashes and other people's data are never included.
type AccountExport struct {
	ExportedAt time.Time          `json:"exported_at"`
	Profile    ExportedProfile    `json:"profile"`
	Rooms      []ExportedRoom     `json:"rooms"`
	Messages   []ExportedMessage  `json:"messages"`
	Reactions  []ExportedReaction `json:"reactions"`
	Reports    []ExportedReport   `json:"reports_filed"`
}

type ExportedProfile struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	LastName   *string    `json:"last_name"`
	Username   string     `json:"username"`
	Email      string     `json:"email"`
	Phone      *string    `json:"phone"`
	AvatarURL  *string    `json:"avatar_url"`
	Role       string     `json:"role"`
	CreatedAt  time.Time  `json:"created_at"`
	LastAccess *time.Time `json:"last_access"`
}

type ExportedRoom struct {
	RoomID   string    `json:"room_id"`
	Name     string    `json:"name"`
	IsDirect bool      `json:"is_direct"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type ExportedMessage struct {
	ID        string     `json:"id"`
	RoomID    string     `json:"room_id"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"created_at"`
	EditedAt  *time.Time `json:"edited_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type ExportedReaction struct {
	MessageID string    `json:"message_id"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}

type ExportedReport struct {
	MessageID string    `json:"message_id"`
	Reason    string    `json:"reason"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
