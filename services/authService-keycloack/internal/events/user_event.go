package events

import "github.com/google/uuid"

const (
	UserCreated = "UserCreated"
	UserUpdated = "UserUpdated"
	UserDeleted = "UserDeleted"
)

type UserCreatedPayload struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}
type UserUpdatedPayload struct {
	UserID    uuid.UUID `json:"user_id"`
	AvatarURL string    `json:"file_url"`
	OldURL    string    `json:"old_url"`
}

type UserDeletedPayload struct {
	UserID   uuid.UUID `json:"user_id"`
	ImageURL string    `json:"image_url"`
}
