package db

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/bjarke-xyz/go-monorepo/libs/common/config"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	cache     *cache.Cache
	keyPrefix string
}

func NewRedisCache(cfg *config.Config) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisConnectionString(),
		Password: cfg.RedisPassword,
	})
	mycache := cache.New(&cache.Options{
		Redis:      rdb,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})
	return &RedisCache{
		cache:     mycache,
		keyPrefix: cfg.RedisPrefix,
	}
}

func (r *RedisCache) getKey(key string) string {
	return r.keyPrefix + ":" + key
}

func (r *RedisCache) Set(ctx context.Context, key string, value any, TTL time.Duration) error {
	err := r.cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   r.getKey(key),
		Value: value,
		TTL:   TTL,
	})
	if err != nil {
		log.Printf("cache set with key %v failed: %v", key, err)
	}
	return err
}

func (r *RedisCache) Get(ctx context.Context, key string, value any) error {
	err := r.cache.Get(ctx, r.getKey(key), value)
	if err != nil {
		if !errors.Is(err, cache.ErrCacheMiss) {
			log.Printf("cache get with key %v failed: %v", key, err)
		}
	}
	return err
}
