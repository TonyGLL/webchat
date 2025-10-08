package application

// CreateMessageDTO is the data transfer object for creating a message.
// The AuthorID is not part of the DTO as it should be derived from the
// authenticated user's context (e.g., JWT).
type CreateMessageDTO struct {
	Text      string `json:"text" validate:"required"`
	ChannelID string `json:"channelId" validate:"required"`
}