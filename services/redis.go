package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
)

var (
	Redis *redis.Client = nil
)

// InitializeRedis connects to Redis and verifies the connection
func InitializeRedis() error {
	opt, _ := redis.ParseURL(os.Getenv("REDIS_URL"))
	Redis = redis.NewClient(opt)

	// Test the connection
	ctx := context.Background()
	_, err := Redis.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return nil
}

// SaveAgentState saves any struct to Redis with a given key
func SaveAgentState(ctx context.Context, key string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	err = Redis.Set(ctx, key, jsonData, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to save to Redis: %w", err)
	}

	log.Debug().Str("key", key).Msg("Agent state saved to Redis")
	return nil
}

// LoadAgentState loads data from Redis into the provided struct
func LoadAgentState(ctx context.Context, key string, data interface{}) error {
	jsonData, err := Redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("no data found for key %s", key)
		}
		return fmt.Errorf("failed to load from Redis: %w", err)
	}

	err = json.Unmarshal([]byte(jsonData), data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	log.Debug().Str("key", key).Msg("Agent state loaded from Redis")
	return nil
}

// DeleteAgentState removes data from Redis
func DeleteAgentState(ctx context.Context, key string) error {
	err := Redis.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete from Redis: %w", err)
	}

	log.Debug().Str("key", key).Msg("Agent state deleted from Redis")
	return nil
}
