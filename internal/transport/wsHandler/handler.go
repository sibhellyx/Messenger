package wshandler

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WsServiceInterface interface {
	HandleConnection(userID, uuid string, conn *websocket.Conn, userAgent, ipAddress string) (string, error)
}

type WsHandler struct {
	service  WsServiceInterface
	upgrader websocket.Upgrader
}

func NewWsHandler(service WsServiceInterface) *WsHandler {
	return &WsHandler{
		service: service,
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

	clientID, err := h.service.HandleConnection(
		userId.(string),
		uuid.(string),
		conn,
		c.Request.UserAgent(),
		c.ClientIP(),
	)

	if err != nil {
		c.JSON(500, gin.H{"error": "failed to handle connection"})
		return
	}

	response := map[string]interface{}{
		"type":      "connection_established",
		"client_id": clientID,
		"user_id":   userId.(string),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to marshal response"})
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, responseBytes); err != nil {
		c.JSON(500, gin.H{"error": "failed to send client_id"})
		return
	}

}
