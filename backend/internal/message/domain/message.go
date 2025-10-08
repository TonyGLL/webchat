package domain

import "time"

type Message struct {
	ID        int       `json:"id"`
	Text      string    `json:"text"`
	AuthorID  int       `json:"author_id"`
	ChannelID int       `json:"channel_id"`
	CreatedAt time.Time `json:"created_at"`
}