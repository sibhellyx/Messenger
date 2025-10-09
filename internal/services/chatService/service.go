package chatservice

import (
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
	DirectedChatCreated(firstId, secondId uint) (*entity.Chat, error)
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
	// getting chat participants
	GetChatParticipants(chatID uint) ([]*entity.ChatParticipant, error)

	// check chat exist
	ChatExists(chatID uint) bool
	// check chat for fulling
	CheckAvailibleForAddParticipantToChat(chatID uint) bool
	// check if chat directed
	CheckChatDirected(chatID uint) bool
	// check user for participant
	ParticipantExist(userID, chatID uint) bool
	// check user for exist
	UserExist(userID uint) bool
}

type ChatService struct {
	repository ChatRepositoryInterface
}

func NewChatService(repository ChatRepositoryInterface) *ChatService {
	return &ChatService{
		repository: repository,
	}
}

func (s *ChatService) CreateChat(userID string, req request.CreateChatRequest) (*entity.Chat, error) {
	slog.Debug("start creating chat")
	err := req.Validate()
	if err != nil {
		slog.Error("failed validate request", "error", err)
		return nil, err
	}

	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		slog.Error("failed parse user_id to uint", "user_id", userID)
		return nil, chaterrors.ErrInvalidUser
	}
	time := time.Now()

	// if chat directed need set 2 members for max check whether it has already been created
	maxMembers := 100
	if req.Type != "" && req.Type == entity.ChatTypeDirect {
		maxMembers = 2
		userId, err := strconv.ParseUint(req.Participants[0].ID, 10, 32)
		if err != nil {
			slog.Error("failed parse user_id to uint", "user_id", userID)
			return nil, chaterrors.ErrInvalidUser
		}

		chat, err := s.repository.DirectedChatCreated(uint(id), uint(userId))
		if err != nil {
			return nil, err
		}

		if chat != nil {
			return chat, nil
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
		return nil, err
	}
	// set role of creator and member
	creatorRole := entity.RoleOwner
	memberRole := entity.RoleMember
	// if direct will set admin role for creator and andther member
	if createdChat.Type == entity.ChatTypeDirect {
		creatorRole = entity.RoleAdmin
		memberRole = entity.RoleAdmin
		// also need check if user want create chat with yourself
		userId, err := strconv.ParseUint(req.Participants[0].ID, 10, 32)
		if err != nil {
			slog.Error("failed parse user_id to uint", "user_id", userID)
			return nil, chaterrors.ErrInvalidUser
		}
		if uint(userId) == uint(id) {
			err = s.repository.DeleteChat(createdChat.ID)
			if err != nil {
				return nil, err
			}
			return nil, chaterrors.ErrCreatingChatWithYourself
		}
	}
	slog.Debug("seted role for creator", "creator_id", createdChat.CreatedBy, "creator role", creatorRole)
	// add creator
	err = s.addParticipant(createdChat.ID, createdChat.CreatedBy, creatorRole)
	if err != nil {
		err = s.repository.DeleteChat(createdChat.ID)
		if err != nil {
			return nil, err
		}
		return nil, err
	}

	// add participant
	for _, p := range req.Participants {
		userId, err := strconv.ParseUint(p.ID, 10, 32)
		if err != nil {
			slog.Error("failed parse user_id to uint", "user_id", userID)
			return nil, chaterrors.ErrInvalidUser
		}
		err = s.addParticipant(createdChat.ID, uint(userId), memberRole)
		if err != nil {
			slog.Warn("failed add member", "error", err)
		}
	}

	slog.Debug("creating chat completed", "chat_id", createdChat.ID)
	return createdChat, nil
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
		return chaterrors.ErrInvalidUser
	}

	chatId, err := strconv.ParseUint(req.Id, 10, 32)
	if err != nil {
		slog.Error("failed parse chat_id to uint", "chat_id", req.Id)
		return chaterrors.ErrInvalidChat
	}

	_, err = s.repository.GetChatById(uint(chatId))
	if err != nil {
		slog.Error("failed get chat", "chat_id", chatId)
		return chaterrors.ErrChatNotFound
	}

	can, err := s.repository.UserCanChange(uint(id), uint(chatId))
	if err != nil {
		slog.Error("failed get participant info", "error", err)
		return err
	}
	if !can {
		return chaterrors.ErrNotPermission
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
		return nil, chaterrors.ErrInvalidChat
	}
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		slog.Error("failed parse user_id to uint", "user_id", userID)
		return nil, chaterrors.ErrInvalidUser
	}
	can, err := s.repository.UserCanChange(uint(id), uint(chatId))
	if err != nil {
		slog.Error("failed get participant info", "error", err)
		return nil, err
	}
	if !can {
		return nil, chaterrors.ErrNotPermission
	}

	chat, err := s.repository.GetChatById(uint(chatId))
	if err != nil {
		slog.Error("failed get chat", "chat_id", req.Id)
		return nil, err
	}

	// if chat directed not updated data
	if chat.Type == entity.ChatTypeDirect {
		slog.Error("cant update direct chat", "chat_id", chatId)
		return nil, chaterrors.ErrCantUpdaeteDirect
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
		return nil, chaterrors.ErrInvalidUser
	}
	return s.repository.GetUserChats(uint(id))
}

func (s *ChatService) GetChats() ([]*entity.Chat, error) {
	return s.repository.GetChats()
}

func (s *ChatService) SearchChatsByName(name string) ([]*entity.Chat, error) {
	if strings.TrimSpace(name) == "" {
		return nil, chaterrors.ErrInvalidNameForSearch
	}

	return s.repository.FindChatsByName(name)
}

func (s *ChatService) AddParticipant(userID string, req request.ParticipantAddRequest) error {
	slog.Debug("add participant to chat", "chat_id", req.Id, "adder_id", userID, "new_participant", req.Id)

	// add Validate
	chatId, err := strconv.ParseUint(req.Id, 10, 32)
	if err != nil {
		slog.Error("failed parse chat_id to uint", "chat_id", req.Id)
		return chaterrors.ErrInvalidChat
	}

	newUser, err := strconv.ParseUint(req.NewUserId, 10, 32)
	if err != nil {
		slog.Error("failed parse user_id to uint", "new_user_id", req.NewUserId)
		return chaterrors.ErrInvalidIdNewParticipant
	}

	err = s.addParticipant(uint(chatId), uint(newUser), entity.RoleMember)
	if err != nil {
		slog.Warn("failed add member", "error", err)
		return err
	}

	return nil
}

func (s *ChatService) addParticipant(chatID, userID uint, role entity.ParticipantRole) error {
	//check user exist
	if !s.repository.UserExist(userID) {
		slog.Warn("user not found", "user_id", userID)
		return chaterrors.ErrUserNotFound
	}
	// check chat exist
	if !s.repository.ChatExists(chatID) {
		slog.Warn("chat not found", "chat_id", chatID)
		return chaterrors.ErrChatNotFound
	}
	// check availible for add participant to chat
	if !s.repository.CheckAvailibleForAddParticipantToChat(chatID) {
		//directed or full
		if !s.repository.CheckChatDirected(chatID) {
			slog.Warn("chat is directed", "chat_id", chatID)
			return chaterrors.ErrChatIsDirected
		}
		slog.Warn("chat is full", "chat_id", chatID)
		return chaterrors.ErrFullChat
	}
	// this user already participant of this chat
	if s.repository.ParticipantExist(userID, chatID) {
		slog.Warn("user already participant this chat", "user_id", userID, "chat_id", chatID)
		return chaterrors.ErrAlreadyParticipant
	}

	participant := entity.ChatParticipant{
		UserID: userID,
		ChatID: chatID,
		Role:   role,
	}

	return s.repository.AddParticipant(participant)
}

func (s *ChatService) GetChatParticipants(chatID string) ([]*entity.ChatParticipant, error) {
	slog.Debug("getting chat participants", "chat_id", chatID)
	chatId, err := strconv.ParseUint(chatID, 10, 32)
	if err != nil {
		slog.Error("failed parse chat_id to uint", "chat_id", chatID)
		return nil, chaterrors.ErrInvalidChat
	}
	participants, err := s.repository.GetChatParticipants(uint(chatId))
	if err != nil {
		return nil, chaterrors.ErrFailedGetParticipants
	}

	slog.Debug("participants sucsessfuly get from chat", "chat_id", chatId, "count participants", len(participants))
	return participants, nil
}
