package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type cache struct {
	client *redis.Client
}

func NewRedis(opts *redis.Options) *cache {
	client := redis.NewClient(opts)

	if err := client.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}

	return &cache{client}
}

func (c *cache) Set(ctx context.Context, key string, value any, ttl time.Duration) (err error) {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *cache) Get(ctx context.Context, key string) (value string, err error) {
	return c.client.Get(ctx, key).Result()
}

func (c *cache) Del(ctx context.Context, keys ...string) (err error) {
	return c.client.Del(ctx, keys...).Err()
}
