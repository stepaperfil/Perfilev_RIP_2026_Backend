package redis

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"

	"tco-backend/internal/app/config"
)

const servicePrefix = "tco_service."

type Client struct {
	cfg    config.RedisConfig
	client *redis.Client
}

func New(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Username:    cfg.User,
		Password:    cfg.Password,
		Addr:        cfg.Host + ":" + strconv.Itoa(cfg.Port),
		DB:          0,
		DialTimeout: cfg.DialTimeout,
		ReadTimeout: cfg.ReadTimeout,
	})

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("cant ping redis: %w", err)
	}

	return &Client{cfg: cfg, client: redisClient}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}
