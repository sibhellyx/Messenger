package chatrepo

import (
	"errors"
	"log/slog"

	"github.com/sibhellyx/Messenger/internal/models/entity"
	"gorm.io/gorm"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{
		db: db,
	}
}

func (r *ChatRepository) CreateChat(chat entity.Chat) (*entity.Chat, error) {
	slog.Debug("creating chat", "creator_id", chat.CreatedBy)

	result := r.db.Create(&chat)
	if result.Error != nil {
		slog.Error("failed create chat", "creator_id", chat.CreatedBy, "error", result.Error)
		return nil, errors.New("failed create chat")
	}

	slog.Info("chat created successfully", "chat_name", chat.Name, "creator_id", chat.CreatedBy)
	return &chat, nil
}

func (r *ChatRepository) AddParticipant(participant entity.ChatParticipant) error {
	slog.Debug("add participant to chat",
		"chat_id", participant.ChatID,
		"user_id", participant.UserID,
		"role", participant.Role,
	)

	result := r.db.Create(&participant)
	if result.Error != nil {
		slog.Error("failed add participant",
			"error", result.Error,
			"chat_id", participant.ChatID,
			"user_id", participant.UserID,
			"role", participant.Role,
		)
		return errors.New("failed add participant")
	}
	slog.Info("add participant successfully",
		"chat_id", participant.ChatID,
		"user_id", participant.UserID,
		"role", participant.Role,
	)
	return nil
}
