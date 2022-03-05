package cache

import "github.com/go-redis/redis/v8"

type Cache struct {
	client redis.Client
}
