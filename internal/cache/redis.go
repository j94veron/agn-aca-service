package cache

import (
	"context"

	"agn-service/internal/config"
	"github.com/redis/go-redis/v9"
)

func NewRedis(cfg config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Pass,
		DB:       cfg.DB,
	})
}

func Ping(ctx context.Context, r *redis.Client) error {
	return r.Ping(ctx).Err()
}
