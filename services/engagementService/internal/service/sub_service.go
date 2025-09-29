package service

import (
	"context"
	"engagementService/internal/event"
	"engagementService/internal/repository"
	"errors"
	"fmt"
	"github.com/Sayan80bayev/go-project/pkg/messaging"
	"log"
	"time"

	"github.com/google/uuid"

	"engagementService/internal/model"
)

var (
	ErrAlreadyFollowing = errors.New("already following")
	ErrNotFollowing     = errors.New("not following")
)

type SubscriptionService struct {
	repo     repository.SubscriptionRepo
	producer messaging.Producer
}

func NewSubscriptionService(r repository.SubscriptionRepo, p messaging.Producer) *SubscriptionService {
	return &SubscriptionService{
		repo:     r,
		producer: p,
	}
}

func (s *SubscriptionService) Follow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	if followerID == uuid.Nil || followeeID == uuid.Nil {
		return errors.New("invalid ids")
	}
	if followerID == followeeID {
		return errors.New("cannot follow self")
	}

	sub := &model.Subscription{
		ID:         uuid.New(),
		FollowerID: followerID,
		FolloweeID: followeeID,
		Approved:   true,
		CreatedAt:  time.Now().UTC(),
	}

	err := s.repo.Create(ctx, sub)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateSubscription) {

			return ErrAlreadyFollowing
		}
		return fmt.Errorf("repo create: %w", err)
	}

	go func() {
		payload := events.SubscriptionCreatedPayload{
			FollowerID: followerID,
			FolloweeID: followeeID,
			CreatedAt:  sub.CreatedAt.Unix(),
		}
		if perr := s.producer.Produce(context.Background(), events.TopicSubscriptionCreated, payload); perr != nil {

			log.Printf("warn: failed to produce SubscriptionCreated event: %v", perr)
		}
	}()

	return nil
}

func (s *SubscriptionService) Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	if followerID == uuid.Nil || followeeID == uuid.Nil {
		return errors.New("invalid ids")
	}

	if err := s.repo.Delete(ctx, followerID, followeeID); err != nil {
		return fmt.Errorf("repo delete: %w", err)
	}

	go func() {
		payload := events.SubscriptionDeletedPayload{
			FollowerID: followerID,
			FolloweeID: followeeID,
			DeletedAt:  time.Now().UTC().Unix(),
		}
		if perr := s.producer.Produce(context.Background(), events.TopicSubscriptionDeleted, payload); perr != nil {
			log.Printf("warn: failed to produce SubscriptionDeleted event: %v", perr)
		}
	}()

	return nil
}

func (s *SubscriptionService) IsFollowing(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error) {
	return s.repo.IsFollowing(ctx, followerID, followeeID)
}

func (s *SubscriptionService) GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int64) ([]model.Subscription, error) {
	return s.repo.GetFollowers(ctx, userID, limit, offset)
}

func (s *SubscriptionService) GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int64) ([]model.Subscription, error) {
	return s.repo.GetFollowing(ctx, userID, limit, offset)
}

func (s *SubscriptionService) CountFollowers(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.repo.CountFollowers(ctx, userID)
}

func (s *SubscriptionService) CountFollowing(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.repo.CountFollowing(ctx, userID)
}
