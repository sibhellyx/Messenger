package authhandler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sibhellyx/Messenger/internal/models/entity"
	authservice "github.com/sibhellyx/Messenger/internal/services/authService"
)

type AuthHandler struct {
	service *authservice.AuthService
}

func NewAuthHandler(service *authservice.AuthService) *AuthHandler {
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
		"message": "created",
	})

}

func WrapError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": err.Error(),
	})
}
