package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{client: client}, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Keys returns all keys matching the specified pattern
func (c *RedisCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.client.Keys(ctx, pattern).Result()
}

// DeletePattern removes all keys matching the specified pattern
func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	// Get all keys matching the pattern
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	
	// Delete each matching key
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	
	if err := iter.Err(); err != nil {
		return err
	}
	
	return nil
}

func (c *RedisCache) Clear(ctx context.Context) error {
	return c.client.FlushAll(ctx).Err()
}

// Ping checks if the Redis server is alive
func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
} 