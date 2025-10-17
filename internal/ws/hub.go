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

	config config.WsConfig
}

func NewHub(conf config.WsConfig) *Hub {
	return &Hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		config:     conf,
	}
}

func (h *Hub) Run() {
	slog.Debug("hub run")
	for {
		select {
		case client := <-h.Register:
			slog.Info("register", "user_id", client.ID)
			h.Clients[client] = true
		case client := <-h.Unregister:
			slog.Info("unregister", "user_id", client.ID)
			delete(h.Clients, client)
		case message := <-h.Broadcast:
			for client := range h.Clients {
				slog.Info("broadcast", "message", string(message), "recived_id", client.ID)
				select {
				case client.Send <- message:
				default:
					client.Close()
					delete(h.Clients, client)
				}
			}
		}
	}
}
