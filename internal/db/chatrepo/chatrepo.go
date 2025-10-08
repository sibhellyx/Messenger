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

	if !r.chatExists(chatID) {
		slog.Warn("chat not found", "chat_id", chatID)
		return chaterrors.ErrChatNotFound
	}

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
		// check if chat directed
		if !r.checkChatDirected(participant.ChatID) {
			slog.Warn("chat is directed", "chat_id", participant.ChatID)
			return chaterrors.ErrChatIsDirected
		}
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

// fix creating new chat with user after deleted (not created) and add for name chat unique id maybe
func (r *ChatRepository) DirectedChatCreated(firstId, secondId uint) (uint, error) {
	slog.Debug("checking directed chat existence",
		"first_user", firstId,
		"second_user", secondId)

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

func (r *ChatRepository) GetChatById(chatID uint) (*entity.Chat, error) {
	slog.Debug("get chat by id", "chat_id", chatID)
	var chat entity.Chat
	err := r.db.First(&chat, chatID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, chaterrors.ErrChatNotFound
		}
		slog.Error("failed to get chat", "chat_id", chatID, "error", err)
		return nil, chaterrors.ErrFailedGetChat
	}
	return &chat, nil
}

func (r *ChatRepository) UpdateChat(chat *entity.Chat) (*entity.Chat, error) {
	slog.Debug("updating chat", "chat_id", chat.ID)

	result := r.db.Save(chat)
	if result.Error != nil {
		slog.Error("failed to update chat",
			"chat_id", chat.ID,
			"error", result.Error)
		return nil, chaterrors.ErrFailedUpdateChat
	}

	slog.Info("chat updated successfully",
		"chat_id", chat.ID,
		"chat_name", chat.Name)
	return chat, nil

}

func (r *ChatRepository) UserCanChange(userID, chatID uint) (bool, error) {
	var participant entity.ChatParticipant
	err := r.db.
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		First(&participant).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		slog.Error("failed to get chat participant", "chat_id", chatID, "user_id", userID, "error", err)
		return false, chaterrors.ErrFailedGetParticipant
	}

	return participant.Role == entity.RoleOwner || participant.Role == entity.RoleAdmin, nil
}

func (r *ChatRepository) GetUserChats(userID uint) ([]*entity.Chat, error) {
	slog.Debug("getting user chats", "user_id", userID)

	var chats []*entity.Chat

	var chatIDs []uint
	err := r.db.Model(&entity.ChatParticipant{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Pluck("chat_id", &chatIDs).Error

	if err != nil {
		slog.Error("failed to get user chat IDs", "user_id", userID, "error", err)
		return nil, chaterrors.ErrFailedGetChats
	}
	// if user has not chats
	if len(chatIDs) == 0 {
		slog.Debug("user has no chats", "user_id", userID)
		return []*entity.Chat{}, nil
	}
	// get all chats user
	err = r.db.
		Where("id IN (?)", chatIDs).
		Find(&chats).Error

	if err != nil {
		slog.Error("failed to get user chats", "user_id", userID, "error", err)
		return nil, chaterrors.ErrFailedGetChats
	}

	slog.Debug("successfully retrieved user chats",
		"user_id", userID,
		"chat_count", len(chats))
	return chats, nil
}

func (r *ChatRepository) GetChats() ([]*entity.Chat, error) {
	slog.Debug("getting all chats")
	var chats []*entity.Chat

	err := r.db.
		Where("deleted_at IS NULL").
		Find(&chats).Error

	if err != nil {
		slog.Error("failed to get all chats", "error", err)
		return nil, chaterrors.ErrFailedGetChats
	}

	slog.Debug("successfully retrieved all chats",
		"chat_count", len(chats))

	return chats, nil
}

// soon add pagination
func (r *ChatRepository) FindChatsByName(name string) ([]*entity.Chat, error) {
	slog.Debug("searching chats by name", "name", name)

	var chats []*entity.Chat

	err := r.db.
		Where("name LIKE ? AND deleted_at IS NULL", "%"+name+"%").
		Find(&chats).Error

	if err != nil {
		slog.Error("failed to search chats by name", "name", name, "error", err)
		return nil, chaterrors.ErrFailedGetChats
	}

	slog.Debug("successfully searched chats by name",
		"name", name,
		"chat_count", len(chats))
	return chats, nil
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

func (r *ChatRepository) checkChatDirected(chatID uint) bool {
	var typeChat entity.ChatType

	r.db.Model(&entity.Chat{}).Where("id = ?", chatID).Select("type").Scan(&typeChat)

	return typeChat != entity.ChatTypeDirect
}
