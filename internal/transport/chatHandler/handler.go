package chathandler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sibhellyx/Messenger/internal/models/entity"
	"github.com/sibhellyx/Messenger/internal/models/request"
)

type ChatServiceInterface interface {
	CreateChat(userID string, req request.CreateChatRequest) (*entity.Chat, error)
	DeleteChat(userID string, req request.ChatRequest) error
	UpdateChat(userID string, req request.UpdateChatRequest) (*entity.Chat, error)
	GetChatsUser(userID string) ([]*entity.Chat, error)
	GetChats() ([]*entity.Chat, error)
	SearchChatsByName(name string) ([]*entity.Chat, error)
	AddParticipant(userID string, req request.ParticipantRequest) error
	RemoveParticipant(userID string, req request.ParticipantRequest) error
	UpdateParticipant(userID string, req request.ParticipantUpdateRequest) error
	GetChatParticipants(chatID string) ([]*entity.ChatParticipant, error)
	LeaveFromChat(chatID string, userID string) error
	GetUsers(search string) ([]*entity.User, error)
}

type ChatHandler struct {
	service ChatServiceInterface
}

func NewChatHandler(service ChatServiceInterface) *ChatHandler {
	return &ChatHandler{
		service: service,
	}
}

func (h *ChatHandler) GetUsers(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	// Получаем параметры поиска и пагинации
	search := c.Query("search")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit := 50 // значение по умолчанию
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	users, err := h.service.GetUsers(search)
	if err != nil {
		slog.Error("failed to get users", "error", err, "user_id", userID)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	// Применяем пагинацию
	total := len(users)
	start := offset
	end := offset + limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedUsers := users[start:end]

	c.JSON(200, gin.H{
		"users": paginatedUsers,
		"pagination": gin.H{
			"total":    total,
			"limit":    limit,
			"offset":   offset,
			"has_more": end < total,
		},
	})
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
	chat, err := h.service.CreateChat(userId.(string), req)
	if err != nil {
		WrapError(c, err)
		return
	}
	if chat == nil {
		WrapError(c, errors.New("failed create chat"))
	}
	c.JSON(http.StatusOK, gin.H{
		"chat": chat,
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
	_, exist := c.Get("user_id")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	chats, err := h.service.GetChats()
	if err != nil {
		WrapError(c, err)
		return
	}

	if len(chats) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"chats":   []string{},
			"count":   0,
			"message": "messanger has no chats",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chats": chats,
		"count": len(chats),
	})

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
	name := c.Query("name")

	if name == "" {
		WrapError(c, errors.New("name parameter is required"))
		return
	}

	if len(name) > 100 {
		WrapError(c, errors.New("search query too long"))
		return
	}

	chats, err := h.service.SearchChatsByName(name)
	if err != nil {
		WrapError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chats": chats,
		"count": len(chats),
	})

}

// leave from chat
func (h *ChatHandler) LeaveChat(c *gin.Context) {
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

	err = h.service.LeaveFromChat(req.Id, userId.(string))
	if err != nil {
		WrapError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "leaved from chat",
	})
}

// get members of chat
func (h *ChatHandler) GetChatParticipants(c *gin.Context) {
	chatID := c.Query("id")

	if chatID == "" {
		WrapError(c, errors.New("id of chat required"))
		return
	}

	if len(chatID) > 10 {
		WrapError(c, errors.New("id of chat too long"))
		return
	}

	particpants, err := h.service.GetChatParticipants(chatID)
	if err != nil {
		WrapError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"participants": particpants,
		"count":        len(particpants),
	})

}

// add member to chat
func (h *ChatHandler) AddParticipant(c *gin.Context) {
	userId, exist := c.Get("user_id")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	var req request.ParticipantRequest
	err := json.NewDecoder(c.Request.Body).Decode(&req)
	if err != nil {
		WrapError(c, err)
		return
	}

	err = h.service.AddParticipant(userId.(string), req)
	if err != nil {
		WrapError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": req.UserId,
	})
}

// remove member from chat
func (h *ChatHandler) RemoveParticipant(c *gin.Context) {
	userId, exist := c.Get("user_id")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	var req request.ParticipantRequest
	err := json.NewDecoder(c.Request.Body).Decode(&req)
	if err != nil {
		WrapError(c, err)
		return
	}

	err = h.service.RemoveParticipant(userId.(string), req)
	if err != nil {
		WrapError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":          "removed from chat",
		"chat_id":         req.Id,
		"deleted_user_id": req.UserId,
	})
}

// update role of member chat
func (h *ChatHandler) UpdateParticipant(c *gin.Context) {
	userId, exist := c.Get("user_id")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	var req request.ParticipantUpdateRequest
	err := json.NewDecoder(c.Request.Body).Decode(&req)
	if err != nil {
		WrapError(c, err)
		return
	}

	err = h.service.UpdateParticipant(userId.(string), req)
	if err != nil {
		WrapError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "user role updated",
	})
}

func WrapError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": err.Error(),
	})
}
