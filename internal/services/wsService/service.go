package wsservice

import (
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sibhellyx/Messenger/internal/ws"
)

type WsService struct {
	hub     *ws.Hub
	clients map[string]*ws.Client // userID -> client
	mu      sync.RWMutex
}

func NewWsService(hub *ws.Hub) *WsService {
	return &WsService{
		hub:     hub,
		clients: make(map[string]*ws.Client),
	}
}

func (s *WsService) HandleConnection(userID, uuid string, conn *websocket.Conn, userAgent, ipAddress string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// close old connection
	if existingClient, exists := s.clients[userID]; exists {
		slog.Info("Closing existing connection for user", "user_id", userID)
		s.hub.Unregister <- existingClient
	}

	// creating new client
	client := ws.NewClient(
		userID,
		uuid,
		conn,
		userAgent,
		ipAddress,
		s.hub,
	)

	s.hub.Register <- client
	s.clients[userID] = client

	go client.ReadPump()
	go client.WritePump()

	slog.Info("New WebSocket connection handled",
		"user_id", userID,
		"uuid", uuid,
		"user_agent", userAgent,
		"ip_address", ipAddress,
		"total_connections", len(s.clients))

	return nil
}
