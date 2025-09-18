package api

import (
	"github.com/gin-gonic/gin"
	authhandler "github.com/sibhellyx/Messenger/internal/transport/authHandler"
)

func CreateRoutes(authHandler *authhandler.AuthHandler) *gin.Engine {
	r := gin.Default()

	return r
}
