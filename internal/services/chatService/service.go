package chatservice

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/sibhellyx/Messenger/internal/models/chaterrors"
	"github.com/sibhellyx/Messenger/internal/models/entity"
	"github.com/sibhellyx/Messenger/internal/models/request"
	"github.com/sibhellyx/Messenger/internal/models/wsmsg"
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
	GetChatParticipants(chatID uint, since *time.Time) ([]*entity.ChatParticipant, error)
	// delete user from chat
	DeleteFromChat(chatID, userID uint) error
	// get participant by chat_id and user_id
	GetParticipantByUserIdAndChatId(userID, chatID uint) (*entity.ChatParticipant, error)
	// update participant
	UpdateParticipant(participant *entity.ChatParticipant) error

	// check chat exist
	ChatExists(chatID uint) bool
	// check chat for fulling
	CheckAvailibleForAddParticipantToChat(chatID uint) bool
	// check if chat directed
	CheckChatDirected(chatID uint) bool
	// check user for participant
	ParticipantExist(userID, chatID uint) bool
	// check participant to owner chat
	ParticipantIsOwner(userID, chatID uint) bool
	// check user for exist
	UserExist(userID uint) bool
}

type WsServiceInterface interface {
	BroadcastMessage(msg []byte) error
}

type ChatService struct {
	repository ChatRepositoryInterface
	service    WsServiceInterface
}

func NewChatService(repository ChatRepositoryInterface, service WsServiceInterface) *ChatService {
	return &ChatService{
		repository: repository,
		service:    service,
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

	msg := wsmsg.ParticipantMsg{
		ChatID: uint(chatId),
		Type:   "participant",
		Action: wsmsg.ChatDeleted,
	}

	responseByte, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal message", "chat_id", chatId, "error", err)
	}

	s.service.BroadcastMessage(responseByte)

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

func (s *ChatService) AddParticipant(userID string, req request.ParticipantRequest) error {
	slog.Debug("add participant to chat", "chat_id", req.Id, "adder_id", userID, "new_participant", req.Id)

	err := req.Validate()
	if err != nil {
		slog.Error("failed validate request", "error", err)
		return err
	}

	// add Validate
	chatId, err := strconv.ParseUint(req.Id, 10, 32)
	if err != nil {
		slog.Error("failed parse chat_id to uint", "chat_id", req.Id)
		return chaterrors.ErrInvalidChat
	}

	newUser, err := strconv.ParseUint(req.UserId, 10, 32)
	if err != nil {
		slog.Error("failed parse user_id to uint", "new_user_id", req.UserId)
		return chaterrors.ErrInvalidIdNewParticipant
	}

	err = s.addParticipant(uint(chatId), uint(newUser), entity.RoleMember)
	if err != nil {
		slog.Warn("failed add member", "error", err)
		return err
	}
	return nil
}

func (s *ChatService) EnterToChat(userID string, req request.ChatRequest) error {
	slog.Debug("user enter to chat", "user_id", userID, "chat_id", req.Id)

	err := req.Validate()
	if err != nil {
		slog.Error("failed validate request", "error", err)
		return err
	}
	chatId, err := strconv.ParseUint(req.Id, 10, 32)
	if err != nil {
		slog.Error("failed parse chat_id to uint", "chat_id", req.Id)
		return chaterrors.ErrInvalidChat
	}

	userId, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		slog.Error("failed parse user_id to uint", "user_id", userID)
		return chaterrors.ErrInvalidUser
	}
	if !s.repository.UserExist(uint(userId)) {
		slog.Warn("user not found", "user_id", userID)
		return chaterrors.ErrUserNotFound
	}
	chat, err := s.repository.GetChatById(uint(chatId))
	if err != nil || chat == nil {
		slog.Error("failed found chat with this id", "chat_id", chat.ID)
		return chaterrors.ErrChatNotFound
	}
	if chat.Type == entity.ChatTypeDirect {
		slog.Error("user can't enter to direct chat", "user_id", userID, "chat_id", chat.ID)
		return chaterrors.ErrChatIsDirected
	}
	if chat.IsPrivate {
		slog.Error("user can't enter to private chat", "user_id", userID, "chat_id", chat.ID)
		return chaterrors.ErrChatIsPrivateEnter
	}
	if !s.repository.CheckAvailibleForAddParticipantToChat(chat.ID) {
		slog.Warn("chat is full", "chat_id", chatId)
		return chaterrors.ErrFullChat
	}

	if s.repository.ParticipantExist(uint(userId), uint(chatId)) {
		slog.Error("user can't enter to chat because already participant", "user_id", userID, "chat_id", chat.ID, "err", err)
		return chaterrors.ErrAlreadyParticipant
	}

	participant := entity.ChatParticipant{
		UserID: uint(userId),
		ChatID: uint(chatId),
		Role:   entity.RoleMember,
	}
	err = s.repository.AddParticipant(participant)
	if err != nil {
		return err
	}
	msg := wsmsg.ParticipantMsg{
		ChatID: uint(chatId),
		UserID: uint(userId),
		Type:   "participant",
		Action: wsmsg.Entered,
	}

	responseByte, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal message", "chat_id", req.Id, "user_id", userId, "error", err)
	}

	s.service.BroadcastMessage(responseByte)
	return nil
}

func (s *ChatService) RemoveParticipant(userID string, req request.ParticipantRequest) error {
	slog.Debug("remove participant from chat", "chat_id", req.Id, "user_id", userID, "participant_id(for delete)", req.UserId)

	err := req.Validate()
	if err != nil {
		slog.Error("failed validate request", "error", err)
		return err
	}

	chatId, err := strconv.ParseUint(req.Id, 10, 32)
	if err != nil {
		slog.Error("failed parse chat_id to uint", "chat_id", req.Id)
		return chaterrors.ErrInvalidChat
	}

	userId, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		slog.Error("failed parse user_id to uint", "user_id", userID)
		return chaterrors.ErrInvalidUser
	}

	participantId, err := strconv.ParseUint(req.UserId, 10, 32)
	if err != nil {
		slog.Error("failed parse user_id(participant for delete) to uint", "new_user_id", req.UserId)
		return chaterrors.ErrInvalidUser
	}

	if !s.repository.UserExist(uint(userId)) {
		slog.Warn("user not found", "user_id", userID)
		return chaterrors.ErrUserNotFound
	}
	if !s.repository.UserExist(uint(participantId)) {
		slog.Warn("user(participant for delete) not found", "user_id", userID)
		return chaterrors.ErrUserNotFound
	}
	if !s.repository.ChatExists(uint(chatId)) {
		slog.Warn("chat not found", "chat_id", chatId)
		return chaterrors.ErrChatNotFound
	}

	user, err := s.repository.GetParticipantByUserIdAndChatId(uint(userId), uint(chatId))
	if err != nil {
		slog.Error("failed get participant(who want remove) info", "error", err)
		return err
	}
	userForRemove, err := s.repository.GetParticipantByUserIdAndChatId(uint(participantId), uint(chatId))
	if err != nil {
		slog.Error("failed get participant(which will delete) info", "error", err)
		return err
	}
	if user.Role == entity.RoleMember {
		slog.Error("failed remove user from chat, member can't remove another users", "chat_id", req.Id, "user_id", userID, "participant_id(for delete)", req.UserId)
		return chaterrors.ErrNotPermission
	}
	if user.Role == entity.RoleAdmin && (userForRemove.Role == entity.RoleAdmin || userForRemove.Role == entity.RoleOwner) {
		slog.Error("failed remove user from chat, admin can't remove another admins or owner", "chat_id", req.Id, "user_id", userID, "participant_id(for delete)", req.UserId)
		return chaterrors.ErrFailedRemoveAdminOrOwnerByAdmin
	}

	err = s.repository.DeleteFromChat(uint(chatId), uint(participantId))
	if err != nil {
		return err
	}

	msg := wsmsg.ParticipantMsg{
		ChatID: uint(chatId),
		UserID: uint(userId),
		Type:   "participant",
		Action: wsmsg.Removed,
	}

	responseByte, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal message", "chat_id", chatId, "user_id", userId, "error", err)
	}

	s.service.BroadcastMessage(responseByte)
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

	err := s.repository.AddParticipant(participant)
	if err != nil {
		return err
	}
	msg := wsmsg.ParticipantMsg{
		ChatID: uint(chatID),
		UserID: uint(userID),
		Type:   "participant",
		Action: wsmsg.Add,
	}
	responseByte, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal message", "chat_id", chatID, "user_id", userID, "error", err)
	}
	s.service.BroadcastMessage(responseByte)
	return nil
}

func (s *ChatService) GetChatParticipants(chatID, sinceParam string) ([]*entity.ChatParticipant, error) {
	var since *time.Time
	if sinceParam != "" {
		parsedSince, parseErr := time.Parse(time.RFC3339, sinceParam)
		if parseErr != nil {
			timestamp, parseErr := strconv.ParseInt(sinceParam, 10, 64)
			if parseErr != nil {
				return nil, errors.New("invalid since parameter format, use RFC3339 or Unix timestamp")
			}
			parsedSince = time.Unix(timestamp, 0)
		}

		if parsedSince.After(time.Now()) {
			return nil, errors.New("since parameter cannot be in the future")
		}

		since = &parsedSince
	}
	slog.Debug("getting chat participants", "chat_id", chatID)
	chatId, err := strconv.ParseUint(chatID, 10, 32)
	if err != nil {
		slog.Error("failed parse chat_id to uint", "chat_id", chatID)
		return nil, chaterrors.ErrInvalidChat
	}

	if !s.repository.ChatExists(uint(chatId)) {
		slog.Warn("chat not found", "chat_id", chatId)
		return nil, chaterrors.ErrChatNotFound
	}

	participants, err := s.repository.GetChatParticipants(uint(chatId), since)
	if err != nil {
		return nil, chaterrors.ErrFailedGetParticipants
	}

	slog.Debug("participants sucsessfuly get from chat", "chat_id", chatId, "count participants", len(participants))
	return participants, nil
}

func (s *ChatService) LeaveFromChat(chatID string, userID string) error {
	slog.Debug("user leave chat", "chat_id", chatID, "user_id", userID)
	chatId, err := strconv.ParseUint(chatID, 10, 32)
	if err != nil {
		slog.Error("failed parse chat_id to uint", "chat_id", chatID)
		return chaterrors.ErrInvalidChat
	}
	userId, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		slog.Error("failed parse user_id to uint", "user_id", userID)
		return chaterrors.ErrInvalidUser
	}

	if !s.repository.ChatExists(uint(chatId)) {
		slog.Warn("chat not found", "chat_id", chatId)
		return chaterrors.ErrChatNotFound
	}

	if !s.repository.UserExist(uint(userId)) {
		slog.Warn("user not found", "user_id", userID)
		return chaterrors.ErrUserNotFound
	}
	// check if user not participant
	if !s.repository.ParticipantExist(uint(userId), uint(chatId)) {
		slog.Warn("user not participant this chat", "user_id", userID, "chat_id", chatID)
		return chaterrors.ErrNotParticipant
	}
	// check role to owner
	if s.repository.ParticipantIsOwner(uint(userId), uint(chatId)) {
		slog.Warn("user is owner this chat", "user_id", userID, "chat_id", chatID)
		return chaterrors.ErrUserIsOwner
	}
	//
	if !s.repository.CheckChatDirected(uint(chatId)) {
		slog.Warn("chat is directed, can't leave from directed chat", "chat_id", chatID)
		return chaterrors.ErrFailedLeaveDirectedChat
	}
	slog.Debug("user sucsessfuly leaved from chat", "chat_id", chatID, "user_id", userID)

	err = s.repository.DeleteFromChat(uint(chatId), uint(userId))
	if err != nil {
		return err
	}

	msg := wsmsg.ParticipantMsg{
		ChatID: uint(chatId),
		UserID: uint(userId),
		Type:   "participant",
		Action: wsmsg.Leaved,
	}

	responseByte, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal message", "chat_id", chatId, "user_id", userId, "error", err)
	}

	s.service.BroadcastMessage(responseByte)

	return nil
}

func (s *ChatService) UpdateParticipant(userID string, req request.ParticipantUpdateRequest) error {
	slog.Debug("update participant", "chat_id", req.Id, "by user_id", userID, "to user_id", req.UserId, "role", req.Role)
	err := req.Validate()
	if err != nil {
		slog.Error("failed validate request", "error", err)
		return err
	}

	role, err := entity.GetRoleForUpdate(req.Role)
	if err != nil {
		slog.Error("failed get role for update", "role from req", req.Role)
		return err
	}

	chatId, err := strconv.ParseUint(req.Id, 10, 32)
	if err != nil {
		slog.Error("failed parse chat_id to uint", "chat_id", req.Id)
		return chaterrors.ErrInvalidChat
	}

	userId, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		slog.Error("failed parse user_id to uint", "user_id", userID)
		return chaterrors.ErrInvalidUser
	}

	participantId, err := strconv.ParseUint(req.UserId, 10, 32)
	if err != nil {
		slog.Error("failed parse user_id(participant for delete) to uint", "new_user_id", req.UserId)
		return chaterrors.ErrInvalidUser
	}

	if !s.repository.UserExist(uint(userId)) {
		slog.Warn("user not found", "user_id", userID)
		return chaterrors.ErrUserNotFound
	}
	if !s.repository.UserExist(uint(participantId)) {
		slog.Warn("user(participant for delete) not found", "user_id", userID)
		return chaterrors.ErrUserNotFound
	}
	if !s.repository.ChatExists(uint(chatId)) {
		slog.Warn("chat not found", "chat_id", chatId)
		return chaterrors.ErrChatNotFound
	}

	user, err := s.repository.GetParticipantByUserIdAndChatId(uint(userId), uint(chatId))
	if err != nil {
		slog.Error("failed get participant(who want update) info", "error", err)
		return err
	}

	if user.Role == entity.RoleMember || user.Role == entity.RoleAdmin {
		slog.Error("failed update user from chat, member or admin can't update another users", "chat_id", req.Id, "user_id", userID, "participant_id(for update)", req.UserId)
		return chaterrors.ErrNotPermission
	}

	userUpdate, err := s.repository.GetParticipantByUserIdAndChatId(uint(participantId), uint(chatId))
	if err != nil {
		slog.Error("failed get participant(who will update) info", "error", err)
		return err
	}
	userUpdate.Role = role
	return s.repository.UpdateParticipant(userUpdate)
}
