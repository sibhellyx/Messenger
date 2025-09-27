package chatrepo

import (
	"errors"
	"log/slog"

	"github.com/sibhellyx/Messenger/internal/models/chaterrors"
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
		return nil, chaterrors.ErrFailedCreateChat
	}

	slog.Info("chat created successfully", "chat_name", chat.Name, "creator_id", chat.CreatedBy)
	return &chat, nil
}

func (r *ChatRepository) DeleteChat(chatID uint) error {
	slog.Debug("deleting chat", "chat_id", chatID)
	err := r.deleteAllParticipantsFromChat(chatID)
	if err != nil {
		return err
	}

	result := r.db.Where("id = ?", chatID).Delete(&entity.Chat{})
	if result.Error != nil {
		slog.Error("failed to delete chat",
			"chat_id", chatID,
			"error", result.Error)
		return chaterrors.ErrDeletingChat
	}

	slog.Info("chat successfully deleted",
		"chat_id", chatID)

	return nil
}

func (r *ChatRepository) AddParticipant(participant entity.ChatParticipant) error {
	slog.Debug("add participant to chat",
		"chat_id", participant.ChatID,
		"user_id", participant.UserID,
		"role", participant.Role,
	)
	if !r.userExist(participant.UserID) {
		slog.Warn("user not found", "user_id", participant.UserID)
		return chaterrors.ErrUserNotFound
	}

	if !r.chatExists(participant.ChatID) {
		slog.Warn("chat not found", "chat_id", participant.ChatID)
		return chaterrors.ErrChatNotFound
	}

	if !r.checkAvailibleForAddParticipantToChat(participant.ChatID) {
		slog.Warn("chat is full", "chat_id", participant.ChatID)
		return chaterrors.ErrFullChat
	}

	if r.participantExist(participant.UserID, participant.ChatID) {
		slog.Warn("user already participant this chat", "user_id", participant.UserID, "chat_id", participant.ChatID)
		return chaterrors.ErrAlreadyParticipant
	}

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

func (r *ChatRepository) deleteAllParticipantsFromChat(chatID uint) error {
	slog.Debug("deleting all participants from chat", "chat_id", chatID)

	result := r.db.Where("chat_id = ?", chatID).Delete(&entity.ChatParticipant{})
	if result.Error != nil {
		slog.Error("failed to delete participants from chat",
			"chat_id", chatID,
			"error", result.Error)
		return chaterrors.ErrDeletingAllParticipants
	}

	slog.Info("successfully deleted participants from chat",
		"chat_id", chatID,
		"deleted_count", result.RowsAffected)
	return nil
}

func (r *ChatRepository) DirectedChatCreated(firstId, secondId uint) (uint, error) {
	slog.Debug("checking directed chat existence",
		"first_user", firstId,
		"second_user", secondId)

	if !r.userExist(firstId) || !r.userExist(secondId) {
		return 0, chaterrors.ErrCheckTwoUsersNotFound
	}

	var result struct {
		ChatID uint `gorm:"column:id"`
	}

	err := r.db.Table("chats").
		Select("chats.id").
		Joins("JOIN chat_participants cp ON cp.chat_id = chats.id").
		Where("chats.type = ?", entity.ChatTypeDirect).
		Where("cp.user_id IN (?, ?)", firstId, secondId).
		Group("chats.id").
		Having("COUNT(DISTINCT cp.user_id) = 2").
		Scan(&result).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, chaterrors.ErrFailedCheckDirectedChat
	}

	return result.ChatID, nil
}

func (r *ChatRepository) userExist(userID uint) bool {
	var count int64
	r.db.Model(&entity.User{}).Where("id = ?", userID).Count(&count)
	return count > 0
}

func (r *ChatRepository) participantExist(userID, chatID uint) bool {
	var count int64
	r.db.Model(&entity.ChatParticipant{}).Where("user_id = ? AND chat_id = ?", userID, chatID).Count(&count)
	return count > 0
}

func (r *ChatRepository) chatExists(chatID uint) bool {
	var count int64
	r.db.Model(&entity.Chat{}).Where("id = ?", chatID).Count(&count)
	return count > 0
}

func (r *ChatRepository) checkAvailibleForAddParticipantToChat(chatID uint) bool {
	var maxMembers int
	r.db.Model(&entity.Chat{}).Where("id = ?", chatID).Select("max_members").Scan(&maxMembers)

	var count int64
	r.db.Model(&entity.ChatParticipant{}).Where("chat_id = ?", chatID).Count(&count)

	availibleCount := maxMembers - int(count)
	return availibleCount > 0
}
