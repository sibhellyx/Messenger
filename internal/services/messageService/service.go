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

type MessageRepositoryInterface interface {
	CreateMessage(ctx context.Context, message *entity.Message) error
	UpdateMessageStatus(ctx context.Context, messageID uint, status entity.MessageStatus) error
	GetMessageByID(ctx context.Context, id uint) (*entity.Message, error)
}

type MessageService struct {
	wsService *wsservice.WsService
	producer  *kafka.Producer
	consumer  *kafka.Consumer
	repo      MessageRepositoryInterface
}

func NewMessageService(wsService *wsservice.WsService, producer *kafka.Producer, repo MessageRepositoryInterface) *MessageService {
	return &MessageService{
		wsService: wsService,
		producer:  producer,
		repo:      repo,
	}
}

func (s *MessageService) SetConsumer(consumer *kafka.Consumer) {
	s.consumer = consumer
}

func (s *MessageService) StartConsumer(ctx context.Context) {
	if s.consumer == nil {
		slog.Error("Consumer is not set")
		return
	}
	s.consumer.Start(ctx)
}

func (s *MessageService) StopConsumer() {
	if s.consumer != nil {
		s.consumer.Stop()
	}
}

func (s *MessageService) ProcessKafkaMessage(ctx context.Context, message entity.Message) error {
	slog.Info("Processing message from Kafka",
		"message_id", message.ID,
		"chat_id", message.ChatID,
		"user_id", message.UserID)

	err := s.repo.UpdateMessageStatus(ctx, message.ID, entity.MessageStatusDelivered)
	if err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}

	wsMessage := map[string]interface{}{
		"type":         "new_message",
		"message_id":   message.ID,
		"chat_id":      message.ChatID,
		"user_id":      message.UserID,
		"content":      message.Content,
		"message_type": message.Type,
		"status":       entity.MessageStatusDelivered,
		"client_id":    message.ClientID,
		"timestamp":    message.CreatedAt,
	}

	if message.FileURL != nil {
		wsMessage["file_url"] = *message.FileURL
		wsMessage["file_name"] = message.FileName
		wsMessage["file_size"] = message.FileSize
		wsMessage["mime_type"] = message.MimeType
	}

	messageBytes, err := json.Marshal(wsMessage)
	if err != nil {
		slog.Error("failed to marshal WebSocket message", "err", err, "message", message)
		return errors.New("failed marshal message json to byte")
	}

	if err := s.wsService.BroadcastMessage(messageBytes); err != nil {
		slog.Warn("Failed to broadcast WebSocket message",
			"error", err,
			"chat_id", message.ChatID)
	}

	slog.Info("Message processed successfully",
		"message_id", message.ID,
		"chat_id", message.ChatID,
		"status", entity.MessageStatusDelivered)

	return nil
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

	message := entity.Message{
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

	if err := message.Validate(); err != nil {
		return err
	}

	// write to repos message with status sent(create message)
	err = s.repo.CreateMessage(ctx, &message)
	if err != nil {
		slog.Error("error create message", "chat_id", message.ChatID, "user_id", message.UserID, "err", err)
		return errors.New("failed create message")
	}

	key := fmt.Sprintf("chat_%d", chatID)
	err = s.producer.SendJSON(ctx, key, message)
	if err != nil {
		slog.Error("error send message to Kafka", "chat_id", message.ChatID, "user_id", message.UserID, "err", err)
		return errors.New("failed send message to Kafka")
	}

	slog.Info("Message sent successfully",
		"message_id", message.ID,
		"chat_id", req.ChatID,
		"user_id", userID,
		"client_id", req.ClientID)

	return nil
}
