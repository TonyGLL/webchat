package domain

// DirectoryEntry is a person in the user directory.
type DirectoryEntry struct {
	ID        int     `json:"id"`
	Username  string  `json:"username"`
	Name      string  `json:"name"`
	LastName  *string `json:"last_name"`
	AvatarURL *string `json:"avatar_url"`
	Online    bool    `json:"online"`
}
