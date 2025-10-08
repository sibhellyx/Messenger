package chatservice

import (
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/sibhellyx/Messenger/internal/models/chaterrors"
	"github.com/sibhellyx/Messenger/internal/models/entity"
	"github.com/sibhellyx/Messenger/internal/models/request"
)

type ChatRepositoryInterface interface {
	//creaating new chat
	CreateChat(chat entity.Chat) (*entity.Chat, error)
	// deleting chat
	DeleteChat(chatID uint) error
	// add participant to chat(use in create and can uce for add to private chat from user)
	AddParticipant(participant entity.ChatParticipant) error
	// check chat, it can be created by another user
	DirectedChatCreated(firstId, secondId uint) (uint, error)
	// get chat by id
	GetChatById(chatID uint) (*entity.Chat, error)
	// check role user for changing and deleting chat
	UserCanChange(userID, chatID uint) (bool, error)
	// update information about chat
	UpdateChat(chat *entity.Chat) (*entity.Chat, error)
	// get chats user
	GetUserChats(userID uint) ([]*entity.Chat, error)
	// get all chats
	GetChats() ([]*entity.Chat, error)
	// geting chats by name searching
	FindChatsByName(name string) ([]*entity.Chat, error)
}

type ChatService struct {
	repository ChatRepositoryInterface
}

func NewChatService(repository ChatRepositoryInterface) *ChatService {
	return &ChatService{
		repository: repository,
	}
}

func (s *ChatService) CreateChat(userID string, req request.CreateChatRequest) (uint, error) {
	slog.Debug("start creating chat")
	err := req.Validate()
	if err != nil {
		slog.Error("failed validate request", "error", err)
		return 0, err
	}

	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		slog.Error("failed parse user_id to uint", "user_id", userID)
		return 0, errors.New("invalid user_id")
	}
	time := time.Now()

	// if chat directed need set 2 members for max check whether it has already been created
	maxMembers := 100
	if req.Type != "" && req.Type == "direct" {
		maxMembers = 2
		chatId, err := s.repository.DirectedChatCreated(uint(id), req.Participants[0].ID)
		if err != nil {
			return 0, err
		}
		if chatId != 0 {
			return chatId, nil
		}
	}

	chat := entity.Chat{
		Name:           req.Name,
		Description:    req.Description,
		Type:           req.Type,
		IsPrivate:      req.IsPrivate,
		CreatedBy:      uint(id),
		MaxMembers:     maxMembers,
		LastActivityAt: &time,
	}
	// create chat
	createdChat, err := s.repository.CreateChat(chat)
	if err != nil {
		return 0, err
	}
	// set role of creator and member
	creatorRole := entity.RoleOwner
	memberRole := entity.RoleMember
	// if direct will set admin role for creator and andther member
	if createdChat.Type == entity.ChatTypeDirect {
		creatorRole = entity.RoleAdmin
		memberRole = entity.RoleAdmin
		// also need check if user want create chat with yourself
		if req.Participants[0].ID == uint(id) {
			err = s.repository.DeleteChat(createdChat.ID)
			if err != nil {
				return 0, err
			}
			return 0, chaterrors.ErrCreatingChatWithYourself
		}
	}
	slog.Debug("seted role for creator", "creator_id", createdChat.CreatedBy, "creator role", creatorRole)
	// add creator
	creator := entity.ChatParticipant{
		ChatID: createdChat.ID,
		UserID: createdChat.CreatedBy,
		Role:   creatorRole,
	}
	err = s.repository.AddParticipant(creator)
	if err != nil {
		err = s.repository.DeleteChat(createdChat.ID)
		if err != nil {
			return 0, err
		}
		return 0, err
	}

	// add participant
	for _, p := range req.Participants {
		participant := entity.ChatParticipant{
			ChatID: createdChat.ID,
			UserID: p.ID,
			Role:   memberRole,
		}
		err = s.repository.AddParticipant(participant)
		if err != nil {
			slog.Warn("failed add member", "error", err)
		}
	}

	slog.Debug("creating chat completed", "chat_id", createdChat.ID)
	return createdChat.ID, nil
}

func (s *ChatService) DeleteChat(userID string, req request.ChatRequest) error {
	slog.Debug("chat deleting start")
	err := req.Validate()
	if err != nil {
		slog.Error("failed validate request", "error", err)
		return err
	}
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		slog.Error("failed parse user_id to uint", "user_id", userID)
		return errors.New("invalid user_id")
	}

	chatId, err := strconv.ParseUint(req.Id, 10, 32)
	if err != nil {
		slog.Error("failed parse chat_id to uint", "chat_id", req.Id)
		return errors.New("invalid chat_id")
	}
	can, err := s.repository.UserCanChange(uint(id), uint(chatId))
	if err != nil {
		slog.Error("failed get participant info", "error", err)
		return err
	}
	if !can {
		return errors.New("participant doesn't have permission to delete this chat")
	}

	err = s.repository.DeleteChat(uint(chatId))
	if err != nil {
		slog.Error("failed delete chat", "chat_id", chatId, "deleter_id", userID)
		return err
	}

	slog.Debug("deleting chat completed", "chat_id", chatId)
	return nil
}

func (s *ChatService) UpdateChat(userID string, req request.UpdateChatRequest) (*entity.Chat, error) {
	slog.Debug("chat update started")

	err := req.Validate()
	if err != nil {
		slog.Error("failed validate request", "error", err)
		return nil, err
	}

	chatId, err := strconv.ParseUint(req.Id, 10, 32)
	if err != nil {
		slog.Error("failed parse chat_id to uint", "chat_id", req.Id)
		return nil, errors.New("invalid chat_id")
	}
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		slog.Error("failed parse user_id to uint", "user_id", userID)
		return nil, errors.New("invalid user_id")
	}
	can, err := s.repository.UserCanChange(uint(id), uint(chatId))
	if err != nil {
		slog.Error("failed get participant info", "error", err)
		return nil, err
	}
	if !can {
		return nil, errors.New("participant doesn't have permission to update this chat")
	}

	chat, err := s.repository.GetChatById(uint(chatId))
	if err != nil {
		slog.Error("failed get chat", "chat_id", req.Id)
		return nil, err
	}

	// if chat directed not updated data
	if chat.Type == entity.ChatTypeDirect {
		slog.Error("cant update direct chat", "chat_id", chatId)
		return nil, errors.New("failed update chat, direct chat not updated")
	}

	// update chat from req
	if req.Name != "" {
		chat.Name = req.Name
	}
	if req.Description != nil {
		chat.Description = req.Description
	}
	if req.AvatarURL != nil {
		chat.AvatarURL = req.AvatarURL
	}
	chat.IsPrivate = req.IsPrivate

	updatedChat, err := s.repository.UpdateChat(chat)
	if err != nil {
		return nil, err
	}

	slog.Debug("chat updated successfully", "chat_id", chatId)
	return updatedChat, nil

}

func (s *ChatService) GetChatsUser(userID string) ([]*entity.Chat, error) {
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		slog.Error("failed parse user_id to uint", "user_id", userID)
		return nil, errors.New("invalid user_id")
	}
	return s.repository.GetUserChats(uint(id))
}

func (s *ChatService) GetChats() ([]*entity.Chat, error) {
	return s.repository.GetChats()
}

func (s *ChatService) SearchChatsByName(name string) ([]*entity.Chat, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("invalid searching name")
	}

	return s.repository.FindChatsByName(name)
}
