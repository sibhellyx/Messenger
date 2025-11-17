package wsservice

import (
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sibhellyx/Messenger/internal/ws"
)

type WsService struct {
	hub     *ws.Hub
	clients map[string]*ws.Client // userID -> client
	mu      sync.RWMutex
}

func NewWsService(hub *ws.Hub) *WsService {
	service := &WsService{
		hub:     hub,
		clients: make(map[string]*ws.Client),
	}

	service.StartHealthCheck(30 * time.Second)

	return service
}

func (s *WsService) HandleConnection(userID, uuid string, conn *websocket.Conn, userAgent, ipAddress string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// close old connection
	if existingClient, exists := s.clients[userID]; exists {
		slog.Info("Closing existing connection for user", "user_id", userID)
		existingClient.Close()
		time.Sleep(100 * time.Millisecond)
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

	clientID := client.ID

	slog.Info("New WebSocket connection handled",
		"user_id", userID,
		"client_id", clientID,
		"uuid", uuid,
		"user_agent", userAgent,
		"ip_address", ipAddress,
		"total_connections", len(s.clients))

	return clientID, nil
}

func (s *WsService) BroadcastMessage(msg []byte) error {
	s.hub.Broadcast <- msg

	slog.Debug("Message broadcasted", "message_size", len(msg))
	return nil

}

func (s *WsService) StartHealthCheck(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			s.checkConnectionsHealth()
		}
	}()
}

func (s *WsService) checkConnectionsHealth() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	maxInactivity := 30 * time.Second

	for userID, client := range s.clients {
		if !client.IsActive() {
			continue
		}
		lastActivity := client.GetLastActivity()
		if time.Since(lastActivity) > maxInactivity {
			slog.Info("Closing inactive connection",
				"user_id", userID,
				"client_id", client.ID,
				"inactivity_duration", time.Since(lastActivity))
			go client.Close()
		}
	}
}
