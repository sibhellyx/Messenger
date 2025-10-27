package redispkg

import (
	"context"
	"fmt"
	"time"
)

type RedisRepository struct {
	client *Client
	ctx    context.Context
}

func NewRedisRepository(client *Client) *RedisRepository {
	return &RedisRepository{
		client: client,
		ctx:    context.Background(),
	}
}

func (r *RedisRepository) SaveRegistrationToken(token, tgName string, ttl time.Duration) error {
	key := fmt.Sprintf("registration:%s", token)
	return r.client.client.Set(r.ctx, key, tgName, ttl).Err()
}

func (r *RedisRepository) GetRegistrationToken(token string) (string, error) {
	key := fmt.Sprintf("registration:%s", token)
	return r.client.client.Get(r.ctx, key).Result()
}

func (r *RedisRepository) DeleteRegistrationToken(token string) error {
	key := fmt.Sprintf("registration:%s", token)
	return r.client.client.Del(r.ctx, key).Err()
}
