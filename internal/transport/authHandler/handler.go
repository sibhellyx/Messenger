package authhandler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sibhellyx/Messenger/internal/models/payload"
	"github.com/sibhellyx/Messenger/internal/models/request"
	"github.com/sibhellyx/Messenger/internal/models/response"
)

type AuthServiceInterface interface {
	Logout(userId string, uuid string) error
	RefreshToken(payload payload.PayloadForRefresh, params request.LoginParams) (response.Tokens, error)
	RegisterUser(user request.RegisterRequest) (string, error)
	SignInWithoutCode(user request.LoginRequest, params request.LoginParams) (response.Tokens, error)
	SignIn(user request.LoginRequest, params request.LoginParams) (uint, error)
	VerifyCode(req request.VerifyCodeRequest, params request.LoginParams) (response.Tokens, error)
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
	var user request.RegisterRequest
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil {
		WrapError(c, err)
		return
	}
	link, err := h.service.RegisterUser(user)
	if err != nil {
		WrapError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"link": link,
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

	id, err := h.service.SignIn(req, params)
	if err != nil {
		WrapError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": id,
	})
}

func (h *AuthHandler) VerifyLogin(c *gin.Context) {
	var req request.VerifyCodeRequest

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

	tokens, err := h.service.VerifyCode(req, params)
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

	payload := payload.PayloadForRefresh{
		UserId:       userId.(string),
		Uuid:         uuid.(string),
		RefreshToken: req.RefreshToken,
	}

	tokens, err := h.service.RefreshToken(payload, params)
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
