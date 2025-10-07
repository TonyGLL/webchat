package domain

import "time"

type Message struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	AuthorID  string    `json:"authorId"`
	ChannelID string    `json:"channelId"`
	CreatedAt time.Time `json:"createdAt"`
}