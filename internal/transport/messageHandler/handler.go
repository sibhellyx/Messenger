package messagehandler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sibhellyx/Messenger/internal/models/entity"
	"github.com/sibhellyx/Messenger/internal/models/request"
	messageservice "github.com/sibhellyx/Messenger/internal/services/messageService"
)

type MessageServiceInterface interface {
	SendMessage(ctx context.Context, userID string, req request.CreateMessage) error
	GetMessagesByChatId(userID, chatID string) ([]*entity.Message, error)
}

type MessageHandler struct {
	service MessageServiceInterface
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

func (h *MessageHandler) GetMessages(c *gin.Context) {
	userId, exist := c.Get("user_id")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	chatID := c.Query("id")

	if chatID == "" {
		WrapError(c, errors.New("id of chat required"))
		return
	}

	if len(chatID) > 10 {
		WrapError(c, errors.New("id of chat too long"))
		return
	}

	messages, err := h.service.GetMessagesByChatId(userId.(string), chatID)
	if err != nil {
		WrapError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
	})
}

func WrapError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": err.Error(),
	})
}
