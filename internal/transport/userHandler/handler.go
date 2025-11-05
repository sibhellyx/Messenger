package userhandler

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sibhellyx/Messenger/internal/models/entity"
)

type UserServiceInterface interface {
	GetUsers(search string) ([]*entity.User, error)
}

type UserHandler struct {
	service UserServiceInterface
}

func NewUserHandler(service UserServiceInterface) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	search := c.Query("search")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit := 50
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
