package chathandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ChatServiceInterface interface {
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

}

func (h *ChatHandler) GetChat(c *gin.Context) {

}

func (h *ChatHandler) UpdateChat(c *gin.Context) {

}

func (h *ChatHandler) DeleteChat(c *gin.Context) {

}

// gets chats all or user chats
func (h *ChatHandler) GetChats(c *gin.Context) {

}

func (h *ChatHandler) GetUserChats(c *gin.Context) {

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
