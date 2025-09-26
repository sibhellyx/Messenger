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
	CreateChat(chat entity.Chat) error
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
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		slog.Error("failed parse user_id to uint", "user_id", userID)
		return errors.New("invalid user_id")
	}
	time := time.Now()
	// validate req
	chat := entity.Chat{
		Name:           req.Name,
		Description:    req.Description,
		Type:           req.Type,
		IsPrivate:      req.IsPrivate,
		CreatedBy:      uint(id),
		MaxMembers:     100,
		LastActivityAt: &time,
	}

	err = s.repository.CreateChat(chat)
	if err != nil {
		return err
	}

	slog.Debug("creating chat completed")
	return nil
}
