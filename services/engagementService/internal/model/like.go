package model

import (
	"github.com/google/uuid"
	"time"
)

type Like struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	PostID    uuid.UUID  `json:"post_id"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}
