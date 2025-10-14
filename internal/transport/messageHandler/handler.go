package messagehandler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sibhellyx/Messenger/internal/models/request"
	messageservice "github.com/sibhellyx/Messenger/internal/services/messageService"
)

type MessageHandler struct {
	service *messageservice.MessageService
}

func NewMessageHandler(service *messageservice.MessageService) *MessageHandler {
	return &MessageHandler{
		service: service,
	}
}

func (h *MessageHandler) SendMessage(c *gin.Context) {
	userId, exist := c.Get("user_id")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	var req request.CreateMessage
	err := json.NewDecoder(c.Request.Body).Decode(&req)
	if err != nil {
		WrapError(c, err)
		return
	}
	err = h.service.SendMessage(c.Request.Context(), userId.(string), req)
	if err != nil {
		WrapError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"result": "ok",
	})
}

func WrapError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": err.Error(),
	})
}
