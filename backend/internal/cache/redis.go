package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"wealthos-backend/internal/domain"
)

type CacheService interface {
	GetUserSettings(ctx context.Context, userID string) (*domain.UserSettings, error)
	SetUserSettings(ctx context.Context, userID string, settings *domain.UserSettings) error
	InvalidateUserSettings(ctx context.Context, userID string) error
}

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(redisURL string) *RedisCache {
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		opts = &redis.Options{
			Addr: redisURL,
		}
	}
	opts.MaxRetries = 1
	opts.DialTimeout = 500 * time.Millisecond
	opts.ReadTimeout = 500 * time.Millisecond
	opts.WriteTimeout = 500 * time.Millisecond

	rdb := redis.NewClient(opts)
	// Test ping
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("[Redis] Warning: failed to ping Redis at %s: %v. Caching disabled.", redisURL, err)
		_ = rdb.Close()
		rdb = nil
	} else {
		log.Printf("[Redis] Successfully connected to Redis at %s", redisURL)
	}

	return &RedisCache{
		client: rdb,
		ttl:    24 * time.Hour,
	}
}

func (r *RedisCache) userSettingsKey(userID string) string {
	return fmt.Sprintf("user_settings:%s", userID)
}

func (r *RedisCache) GetUserSettings(ctx context.Context, userID string) (*domain.UserSettings, error) {
	if r == nil || r.client == nil || userID == "" {
		return nil, nil
	}

	val, err := r.client.Get(ctx, r.userSettingsKey(userID)).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	} else if err != nil {
		log.Printf("[Redis] GetUserSettings error for user %s: %v", userID, err)
		return nil, nil // Fallback silently
	}

	var settings domain.UserSettings
	if err := json.Unmarshal([]byte(val), &settings); err != nil {
		return nil, nil
	}

	return &settings, nil
}

func (r *RedisCache) SetUserSettings(ctx context.Context, userID string, settings *domain.UserSettings) error {
	if r == nil || r.client == nil || userID == "" || settings == nil {
		return nil
	}

	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	err = r.client.Set(ctx, r.userSettingsKey(userID), string(data), r.ttl).Err()
	if err != nil {
		log.Printf("[Redis] SetUserSettings error for user %s: %v", userID, err)
	}
	return nil
}

func (r *RedisCache) InvalidateUserSettings(ctx context.Context, userID string) error {
	if r == nil || r.client == nil || userID == "" {
		return nil
	}

	err := r.client.Del(ctx, r.userSettingsKey(userID)).Err()
	if err != nil {
		log.Printf("[Redis] InvalidateUserSettings error for user %s: %v", userID, err)
	}
	return nil
}
