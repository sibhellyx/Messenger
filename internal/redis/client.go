package redispkg

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
	"github.com/sibhellyx/Messenger/internal/config"
)

type Client struct {
	client *redis.Client
	config config.RedisConfig
}

func NewClient(cfg config.RedisConfig) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &Client{
		client: rdb,
		config: cfg,
	}
}

func (c *Client) Connect(ctx context.Context) error {
	_, err := c.client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	slog.Info("successfully connected to Redis")
	return nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) GetClient() *redis.Client {
	return c.client
}
