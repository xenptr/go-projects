package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xenptr/go-projects/todo-list-api/internal/config"
)

type Client struct {
	Client *redis.Client
}

func New(cfg *config.Config) (*Client, error) {
	opt := &redis.Options{
		Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
		Username: cfg.RedisUsername,
		Password: cfg.RedisPassword,
		DB:       0,

		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Client{
		Client: client,
	}, nil
}

func (c *Client) Close() error {
	return c.Client.Close()
}
