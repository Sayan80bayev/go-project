package caching

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/Sayan80bayev/go-project/pkg/logging"
	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type RedisService struct {
	client *redis.Client
	logger *logrus.Logger
}

// NewRedisService initializes Redis client from config
func NewRedisService(cfg RedisConfig) (*RedisService, error) {
	logger := logging.GetLogger()

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		logger.Errorf("❌ Failed to connect to Redis at %s: %v", cfg.Addr, err)
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logger.Infof("✅ Connected to Redis at %s (DB=%d)", cfg.Addr, cfg.DB)

	return &RedisService{client: client, logger: logger}, nil
}

func (c *RedisService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if err := c.client.Set(ctx, key, value, expiration).Err(); err != nil {
		c.logger.Errorf("Redis SET error for key=%s: %v", key, err)
		return err
	}
	c.logger.Debugf("Redis SET key=%s (exp=%s)", key, expiration)
	return nil
}

func (c *RedisService) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		c.logger.Debugf("Redis GET miss key=%s", key)
		return "", nil
	} else if err != nil {
		c.logger.Errorf("Redis GET error for key=%s: %v", key, err)
		return "", err
	}
	c.logger.Debugf("Redis GET hit key=%s", key)
	return val, nil
}

func (c *RedisService) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		c.logger.Errorf("Redis DEL error for key=%s: %v", key, err)
		return err
	}
	c.logger.Debugf("Redis DEL key=%s", key)
	return nil
}

func (c *RedisService) Publish(ctx context.Context, channel, message string) error {
	if err := c.client.Publish(ctx, channel, message).Err(); err != nil {
		c.logger.Errorf("Redis PUBLISH error on channel=%s: %v", channel, err)
		return err
	}
	c.logger.Debugf("Redis PUBLISH channel=%s message=%s", channel, message)
	return nil
}

func (c *RedisService) Exists(ctx context.Context, key string) (bool, error) {
	res, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		c.logger.Errorf("Redis EXISTS error for key=%s: %v", key, err)
		return false, err
	}
	return res > 0, nil
}

func (c *RedisService) Subscribe(ctx context.Context, channel string) *redis.PubSub {
	c.logger.Infof("Redis SUBSCRIBE channel=%s", channel)
	return c.client.Subscribe(ctx, channel)
}

// Close gracefully closes Redis connection
func (c *RedisService) Close() error {
	if err := c.client.Close(); err != nil {
		c.logger.Errorf("Error closing Redis client: %v", err)
		return err
	}
	c.logger.Info("Redis client closed gracefully")
	return nil
}
