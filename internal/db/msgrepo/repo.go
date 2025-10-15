package msgrepo

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/sibhellyx/Messenger/internal/models/entity"
	"gorm.io/gorm"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{
		db: db,
	}
}

func (r *MessageRepository) CreateMessage(ctx context.Context, message *entity.Message) error {
	slog.Debug("creating message", "chat_id", message.ChatID, "user_id", message.UserID)
	now := time.Now()
	message.CreatedAt = now
	message.UpdatedAt = now

	return r.db.WithContext(ctx).Create(message).Error
}

func (r *MessageRepository) UpdateMessageStatus(ctx context.Context, messageID uint, status entity.MessageStatus) error {
	slog.Debug("update status message", "message_id", messageID, "status", status)
	err := r.db.WithContext(ctx).Model(&entity.Message{}).
		Where("id = ?", messageID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error

	if err != nil {
		slog.Error("error update status message", "message_id", messageID, "err", err)
		return errors.New("failed update message status")
	}
	return err
}

func (r *MessageRepository) GetMessageByID(ctx context.Context, id uint) (*entity.Message, error) {
	slog.Debug("get message by id", "message_id", id)
	var message entity.Message
	err := r.db.WithContext(ctx).First(&message, id).Error
	if err != nil {
		slog.Error("error failed get message", "message_id", id, "err", err)
		return &message, errors.New("failed get message")
	}
	return &message, err
}
