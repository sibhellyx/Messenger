package redispkg

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sibhellyx/Messenger/internal/models/entity"
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

func (r *RedisRepository) SaveRegistrationToken(token string, user entity.User, ttl time.Duration) error {
	key := fmt.Sprintf("registration:%s", token)

	// Сериализуем пользователя в JSON
	userJSON, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user to JSON: %w", err)
	}

	return r.client.client.Set(r.ctx, key, userJSON, ttl).Err()
}

func (r *RedisRepository) GetRegistrationToken(token string) (entity.User, error) {
	key := fmt.Sprintf("registration:%s", token)

	data, err := r.client.client.Get(r.ctx, key).Result()
	if err != nil {
		return entity.User{}, err
	}

	var user entity.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return entity.User{}, fmt.Errorf("failed to unmarshal user from JSON: %w", err)
	}

	return user, nil
}

func (r *RedisRepository) DeleteRegistrationToken(token string) error {
	key := fmt.Sprintf("registration:%s", token)
	return r.client.client.Del(r.ctx, key).Err()
}

func (r *RedisRepository) SaveLoginCode(userID uint, code string, ttl time.Duration) error {
	key := fmt.Sprintf("login_code:%d", userID)

	codeData := map[string]interface{}{
		"code":       code,
		"created_at": time.Now(),
		"attempts":   0,
	}

	jsonData, err := json.Marshal(codeData)
	if err != nil {
		return fmt.Errorf("failed to marshal code data: %w", err)
	}

	return r.client.client.Set(r.ctx, key, jsonData, ttl).Err()
}

func (r *RedisRepository) GetLoginCode(userID uint) (string, error) {
	key := fmt.Sprintf("login_code:%d", userID)

	data, err := r.client.client.Get(r.ctx, key).Result()
	if err != nil {
		return "", err
	}

	var codeData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &codeData); err != nil {
		return "", fmt.Errorf("failed to unmarshal code data: %w", err)
	}

	code, ok := codeData["code"].(string)
	if !ok {
		return "", fmt.Errorf("invalid code format")
	}

	return code, nil
}

func (r *RedisRepository) IncrementLoginAttempts(userID uint) error {
	key := fmt.Sprintf("login_code:%d", userID)

	data, err := r.client.client.Get(r.ctx, key).Result()
	if err != nil {
		return err
	}

	var codeData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &codeData); err != nil {
		return fmt.Errorf("failed to unmarshal code data: %w", err)
	}

	attempts, _ := codeData["attempts"].(float64)
	codeData["attempts"] = attempts + 1

	jsonData, err := json.Marshal(codeData)
	if err != nil {
		return fmt.Errorf("failed to marshal code data: %w", err)
	}

	return r.client.client.Set(r.ctx, key, jsonData, time.Minute*10).Err() // Reset TTL
}

func (r *RedisRepository) GetLoginAttempts(userID uint) (int, error) {
	key := fmt.Sprintf("login_code:%d", userID)

	data, err := r.client.client.Get(r.ctx, key).Result()
	if err != nil {
		return 0, err
	}

	var codeData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &codeData); err != nil {
		return 0, fmt.Errorf("failed to unmarshal code data: %w", err)
	}

	attempts, _ := codeData["attempts"].(float64)
	return int(attempts), nil
}

func (r *RedisRepository) SaveUserRegistration(userID uint, tgChatId int64) error {
	key := fmt.Sprintf("user_reg:%d", userID)
	return r.client.client.Set(r.ctx, key, tgChatId, 0).Err()
}

func (r *RedisRepository) GetUserRegistration(userID uint) (int64, error) {
	key := fmt.Sprintf("user_reg:%d", userID)
	result, err := r.client.client.Get(r.ctx, key).Int64()
	if err == redis.Nil {
		return 0, fmt.Errorf("user registration not found for user_id: %d", userID)
	}
	return result, err
}
