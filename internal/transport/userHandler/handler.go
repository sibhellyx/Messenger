package userhandler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sibhellyx/Messenger/internal/models/entity"
	"github.com/sibhellyx/Messenger/internal/models/request"
	"github.com/sibhellyx/Messenger/internal/models/response"
)

type UserServiceInterface interface {
	GetUsers(search string) ([]*entity.User, error)
	GetUsersWithProfiles(search string) ([]*response.UserWithProfile, error)
	UpdateProfile(userId string, req request.ProfileRequest) error
	GetFullInfoAboutUser(userID string) (*response.UserWithProfile, error)
}

type UserHandler struct {
	service UserServiceInterface
}

func NewUserHandler(service UserServiceInterface) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

func (h *UserHandler) UpdateUserProfile(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var req request.ProfileRequest
	err := json.NewDecoder(c.Request.Body).Decode(&req)
	if err != nil {
		WrapError(c, err)
		return
	}

	err = h.service.UpdateProfile(userID.(string), req)
	if err != nil {
		WrapError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"result": "ok",
	})
}

func (h *UserHandler) GetUserProfile(c *gin.Context) {
	_, exist := c.Get("user_id")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	profileUserID := c.Query("user_id")

	userInfo, err := h.service.GetFullInfoAboutUser(profileUserID)
	if err != nil {
		slog.Error("failed to get user profile", "error", err, "user_id", profileUserID)
		WrapError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"profile": userInfo,
	})
}

func (h *UserHandler) GetMyProfile(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	userInfo, err := h.service.GetFullInfoAboutUser(userID.(string))
	if err != nil {
		slog.Error("failed to get my profile", "error", err, "user_id", userID)
		WrapError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"profile": userInfo,
	})
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

func (h *UserHandler) GetUsersWithProfiles(c *gin.Context) {
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

	users, err := h.service.GetUsersWithProfiles(search)
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

func WrapError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": err.Error(),
	})
}
