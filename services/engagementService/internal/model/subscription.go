package model

import (
	"time"

	"github.com/google/uuid"
)

// Subscription describes one user's subscription to another
type Subscription struct {
	ID         uuid.UUID  `json:"id"`
	FollowerID uuid.UUID  `json:"follower_id"` // who subscribes
	FolloweeID uuid.UUID  `json:"followee_id"` // whom they subscribe to
	Approved   bool       `json:"approved"`       // for example, for private accounts
	CreatedAt  time.Time  `json:"created_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"` // soft delete
}