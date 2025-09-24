package wshandler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sibhellyx/Messenger/internal/ws"
)

type WsHandler struct {
	hub      *ws.Hub
	upgrader websocket.Upgrader
}

func NewWsHandler(hub *ws.Hub) *WsHandler {
	return &WsHandler{
		hub: hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

func (h *WsHandler) Connect(c *gin.Context) {
	uuid, exist := c.Get("uuid")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	userId, exist := c.Get("user_id")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to upgrade connection"})
		return
	}

	client := ws.NewClient(
		userId.(string),
		uuid.(string),
		conn,
		c.Request.UserAgent(),
		c.ClientIP(),
		h.hub,
	)

	h.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()

	slog.Info("New client connected",
		"user_id", client.ID,
		"uuid", client.UUID,
		"total_clients", len(h.hub.Clients))

}
