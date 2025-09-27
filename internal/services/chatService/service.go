package chatservice

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/sibhellyx/Messenger/internal/models/entity"
	"github.com/sibhellyx/Messenger/internal/models/request"
)

type ChatRepositoryInterface interface {
	CreateChat(chat entity.Chat) (*entity.Chat, error)
	AddParticipant(participant entity.ChatParticipant) error
}

type ChatService struct {
	repository ChatRepositoryInterface
}

func NewChatService(repository ChatRepositoryInterface) *ChatService {
	return &ChatService{
		repository: repository,
	}
}

func (s *ChatService) CreateChat(userID string, req request.CreateChatRequest) error {
	slog.Debug("start creating chat")
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
	time := time.Now()

	chat := entity.Chat{
		Name:           req.Name,
		Description:    req.Description,
		Type:           req.Type,
		IsPrivate:      req.IsPrivate,
		CreatedBy:      uint(id),
		MaxMembers:     100,
		LastActivityAt: &time,
	}
	// create chat
	createdChat, err := s.repository.CreateChat(chat)
	if err != nil {
		return err
	}
	// set role of creator and member
	creatorRole := entity.RoleOwner
	memberRole := entity.RoleMember
	// if direct will set admin role for creator and andther member
	if createdChat.Type == entity.ChatTypeDirect {
		creatorRole = entity.RoleAdmin
		memberRole = entity.RoleAdmin
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
		return err
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

	slog.Debug("creating chat completed")
	return nil
}
