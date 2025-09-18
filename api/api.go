package api

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/sibhellyx/Messenger/internal/middleware"
	authhandler "github.com/sibhellyx/Messenger/internal/transport/authHandler"
)

func CreateRoutes(authHandler *authhandler.AuthHandler, logger *slog.Logger) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.LoggingMiddleware(logger))

	r.POST("/register", authHandler.Register)

	return r
}
