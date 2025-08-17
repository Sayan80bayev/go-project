package service

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type CacheService interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	Get(ctx context.Context, key string) (string, error)

	Delete(ctx context.Context, key string) error

	Publish(ctx context.Context, channel, message string) error

	Exists(ctx context.Context, key string) (bool, error)

	Subscribe(ctx context.Context, channel string) *redis.PubSub
}
