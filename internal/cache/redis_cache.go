package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	redisKeyPrefix = "BENZINPRISER"
	ErrNil         = errors.New("nil")
)

func NewRedisCache(addr string, username string, password string) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Username: username,
		Password: password,
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			cn.ClientSetName(ctx, "benzinpriser-server")
			return nil
		},
	})

	return &Cache{
		client: *client,
	}
}

func (c *Cache) getKey(key string) string {
	return fmt.Sprintf("%v:%v", redisKeyPrefix, key)
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("could not marshal value: %w", err)
	}
	jsonStr := string(jsonBytes)
	return c.client.Set(ctx, c.getKey(key), jsonStr, expiration).Err()
}
func (c *Cache) Get(ctx context.Context, key string, v interface{}) error {
	jsonStr, err := c.client.Get(ctx, c.getKey(key)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrNil
		} else {
			return fmt.Errorf("error getting value with key %v: %w", key, err)
		}
	}
	jsonBytes := []byte(jsonStr)
	if err := json.Unmarshal(jsonBytes, v); err != nil {
		return fmt.Errorf("could not unmarshal: %w", err)
	}
	return nil
}
