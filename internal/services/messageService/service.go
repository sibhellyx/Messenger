package messageservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/sibhellyx/Messenger/internal/kafka"
	"github.com/sibhellyx/Messenger/internal/models/entity"
	"github.com/sibhellyx/Messenger/internal/models/request"
	wsservice "github.com/sibhellyx/Messenger/internal/services/wsService"
)

type MessageService struct {
	wsService *wsservice.WsService
	producer  *kafka.Producer
}

func NewMessageService(wsService *wsservice.WsService, producer *kafka.Producer) *MessageService {
	return &MessageService{
		wsService: wsService,
		producer:  producer,
	}
}

func (s *MessageService) SendMessage(ctx context.Context, userID string, req request.CreateMessage) error {
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		slog.Error("failed parse user_id to uint", "user_id", userID)
		return errors.New("failed parse user_id")
	}
	chatID, err := strconv.ParseUint(req.ChatID, 10, 32)
	if err != nil {
		slog.Error("failed parse chat_id to uint", "chat_id", userID)
		return errors.New("failed parse chat_id")
	}
	kafkaMessage := entity.Message{
		ChatID:    uint(chatID),
		UserID:    uint(id),
		Type:      req.Type,
		Content:   req.Content,
		Status:    entity.MessageStatusSent,
		ClientID:  req.ClientID,
		FileURL:   req.FileURL,
		FileName:  req.FileName,
		FileSize:  req.FileSize,
		MimeType:  req.MimeType,
		ReplyToID: req.ReplyToID,
	}

	if err := kafkaMessage.Validate(); err != nil {
		return fmt.Errorf("message validation failed: %w", err)
	}

	key := fmt.Sprintf("chat_%d", chatID)

	err = s.producer.SendJSON(ctx, key, kafkaMessage)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	wsMessage := map[string]interface{}{
		"type":      "message",
		"chat_id":   req.ChatID,
		"user_id":   userID,
		"content":   req.Content,
		"client_id": req.ClientID,
		"timestamp": kafkaMessage.CreatedAt,
	}

	wsMessageBytes, err := json.Marshal(wsMessage)
	if err != nil {
		slog.Warn("Failed to marshal WebSocket message", "error", err)
	} else {
		if err := s.wsService.BroadcastMessage(wsMessageBytes); err != nil {
			slog.Warn("Failed to broadcast WebSocket message", "error", err)
		}
	}

	slog.Info("Message sent successfully",
		"chat_id", req.ChatID,
		"user_id", userID,
		"client_id", req.ClientID,
		"content_length", len(req.Content))

	return nil
}
