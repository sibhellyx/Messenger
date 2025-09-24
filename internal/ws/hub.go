package ws

import (
	"log/slog"

	"github.com/sibhellyx/Messenger/internal/config"
)

type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client

	logger *slog.Logger
	config config.WsConfig
}

func NewHub(logger *slog.Logger, conf config.WsConfig) *Hub {
	return &Hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		logger:     logger,
		config:     conf,
	}
}

func (h *Hub) Run() {
	h.logger.Debug("hub run")
	for {
		select {
		case client := <-h.Register:
			h.logger.Info("register", "user_id", client.ID)
			h.Clients[client] = true
		case client := <-h.Unregister:
			h.logger.Info("unregister", "user_id", client.ID)
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
		case message := <-h.Broadcast:
			for client := range h.Clients {
				h.logger.Info("broadcast", "message", string(message), "recived_id", client.ID)
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}
