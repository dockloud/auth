package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewCache(dsn string) (*redis.Client, error) {
	ctx := context.Background()

	opts, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to Redis: %w", err)
	}

	opts.DialTimeout = 5 * time.Second
	opts.ReadTimeout = 3 * time.Second
	opts.WriteTimeout = 3 * time.Second
	opts.MaxRetries = 3

	red := redis.NewClient(opts)

	_, err = red.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("error connecting to Redis: %w", err)
	}
	return red, nil
}
