package chathandler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sibhellyx/Messenger/internal/models/entity"
	"github.com/sibhellyx/Messenger/internal/models/request"
)

type ChatServiceInterface interface {
	CreateChat(userID string, req request.CreateChatRequest) (uint, error)
	DeleteChat(userID string, req request.ChatRequest) error
	UpdateChat(userID string, req request.UpdateChatRequest) (*entity.Chat, error)
	GetChatsUser(userID string) ([]*entity.Chat, error)
}

type ChatHandler struct {
	service ChatServiceInterface
}

func NewChatHandler(service ChatServiceInterface) *ChatHandler {
	return &ChatHandler{
		service: service,
	}
}

// simple crud for chat
func (h *ChatHandler) CreateChat(c *gin.Context) {
	userId, exist := c.Get("user_id")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	var req request.CreateChatRequest
	err := json.NewDecoder(c.Request.Body).Decode(&req)
	if err != nil {
		WrapError(c, err)
		return
	}
	id, err := h.service.CreateChat(userId.(string), req)
	if err != nil {
		WrapError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chat_id": strconv.FormatUint(uint64(id), 10),
	})

}

func (h *ChatHandler) UpdateChat(c *gin.Context) {
	userId, exist := c.Get("user_id")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	var req request.UpdateChatRequest
	err := json.NewDecoder(c.Request.Body).Decode(&req)
	if err != nil {
		WrapError(c, err)
		return
	}

	updatedChat, err := h.service.UpdateChat(userId.(string), req)
	if err != nil {
		WrapError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "chat updated",
		"chat":   updatedChat,
	})
}

func (h *ChatHandler) DeleteChat(c *gin.Context) {
	userId, exist := c.Get("user_id")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	var req request.ChatRequest
	err := json.NewDecoder(c.Request.Body).Decode(&req)
	if err != nil {
		WrapError(c, err)
		return
	}

	err = h.service.DeleteChat(userId.(string), req)
	if err != nil {
		WrapError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "chat deleted",
	})
}

// gets chats all or user chats
func (h *ChatHandler) GetChats(c *gin.Context) {

}

func (h *ChatHandler) GetUserChats(c *gin.Context) {
	userId, exist := c.Get("user_id")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	chats, err := h.service.GetChatsUser(userId.(string))
	if err != nil {
		WrapError(c, err)
		return
	}
	if len(chats) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"chats":   []string{},
			"count":   0,
			"message": "user has no chats",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chats": chats,
		"count": len(chats),
	})
}

func (h *ChatHandler) FindChats(c *gin.Context) {

}

// chat with users actions endpoints
// join to open chat
func (h *ChatHandler) JoinChat(c *gin.Context) {

}

// leave from chat
func (h *ChatHandler) LeaveChat(c *gin.Context) {

}

// get members of chat
func (h *ChatHandler) GetChatParticipants(c *gin.Context) {

}

// add member to chat
func (h *ChatHandler) AddParticipants(c *gin.Context) {

}

// remove member from chat
func (h *ChatHandler) RemoveParticipant(c *gin.Context) {

}

// update role of member chat
func (h *ChatHandler) UpdateParticipantRole(c *gin.Context) {

}

func WrapError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": err.Error(),
	})
}
