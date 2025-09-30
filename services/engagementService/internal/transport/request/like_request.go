package request

import "github.com/google/uuid"

type LikeRequest struct {
	PostID uuid.UUID `json:"post_id" validate:"required,uuid"`
	UserID uuid.UUID `json:"user_id" validate:"required,uuid"`
}
