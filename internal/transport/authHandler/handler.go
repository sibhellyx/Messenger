package authhandler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sibhellyx/Messenger/internal/models/entity"
	"github.com/sibhellyx/Messenger/internal/models/request"
	"github.com/sibhellyx/Messenger/internal/models/response"
)

type AuthServiceInterface interface {
	Logout(userId string, uuid string) error
	RefreshToken(tokens response.Tokens, params request.LoginParams) (response.Tokens, error)
	RegisterUser(user entity.User) error
	SignIn(user request.LoginRequest, params request.LoginParams) (response.Tokens, error)
}

type AuthHandler struct {
	service AuthServiceInterface
}

func NewAuthHandler(service AuthServiceInterface) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var user entity.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil {
		WrapError(c, err)
		return
	}
	err = h.service.RegisterUser(user)
	if err != nil {
		WrapError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"result": "user created",
	})
}

func (h *AuthHandler) SignIn(c *gin.Context) {
	var req request.LoginRequest
	err := json.NewDecoder(c.Request.Body).Decode(&req)
	if err != nil {
		WrapError(c, err)
		return
	}

	userAgent := c.Request.UserAgent()
	ip := c.ClientIP()

	params := request.LoginParams{
		UserAgent: userAgent,
		LastIp:    ip,
	}

	tokens, err := h.service.SignIn(req, params)
	if err != nil {
		WrapError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Tokens{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	access := c.GetHeader("Auth")

	var req request.RefreshTokenRequest
	err := json.NewDecoder(c.Request.Body).Decode(&req)
	if err != nil {
		WrapError(c, err)
		return
	}

	userAgent := c.Request.UserAgent()
	ip := c.ClientIP()

	params := request.LoginParams{
		UserAgent: userAgent,
		LastIp:    ip,
	}

	tokensForRefresh := response.Tokens{
		AccessToken:  access,
		RefreshToken: req.RefreshToken,
	}

	tokens, err := h.service.RefreshToken(tokensForRefresh, params)
	if err != nil {
		WrapError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Tokens{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

func (h *AuthHandler) LogoutUser(c *gin.Context) {
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

	err := h.service.Logout(userId.(string), uuid.(string))
	if err != nil {
		WrapError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "exit successfully",
	})
}

func WrapError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": err.Error(),
	})
}
