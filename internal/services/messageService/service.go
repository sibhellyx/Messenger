package messageservice

import (
	"github.com/sibhellyx/Messenger/internal/models/request"
	wsservice "github.com/sibhellyx/Messenger/internal/services/wsService"
)

type MessageService struct {
	wsService *wsservice.WsService
}

func NewMessageService(wsService *wsservice.WsService) *MessageService {
	return &MessageService{
		wsService: wsService,
	}
}

func (s *MessageService) SendMessage(userID string, req request.CreateMessage) error {
	return s.wsService.BroadcastMessage([]byte(req.Content))
}
