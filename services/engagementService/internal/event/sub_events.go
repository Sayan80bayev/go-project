package events

import "github.com/google/uuid"

const (
	TopicSubscriptionCreated = "subscription.created"
	TopicSubscriptionDeleted = "subscription.deleted"
)

type SubscriptionCreatedPayload struct {
	FollowerID uuid.UUID `json:"follower_id"`
	FolloweeID uuid.UUID `json:"followee_id"`
	CreatedAt  int64     `json:"created_at_unix"`
}

type SubscriptionDeletedPayload struct {
	FollowerID uuid.UUID `json:"follower_id"`
	FolloweeID uuid.UUID `json:"followee_id"`
	DeletedAt  int64     `json:"deleted_at_unix"`
}
