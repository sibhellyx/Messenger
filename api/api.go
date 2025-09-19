package api

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/sibhellyx/Messenger/internal/middleware"
)

type AuthHandlerInterface interface {
	LogoutUser(c *gin.Context)
	RefreshToken(c *gin.Context)
	Register(c *gin.Context)
	SignIn(c *gin.Context)
}

func CreateRoutes(authHandler AuthHandlerInterface, logger *slog.Logger, m middleware.JwtManagerInterface, repo middleware.SessionRepositoryInterface) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.LoggingMiddleware(logger))

	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.SignIn)
	r.POST("/refresh", middleware.AuthMiddleware(m, repo), authHandler.RefreshToken)
	r.POST("/logout", middleware.AuthMiddleware(m, repo), authHandler.LogoutUser)

	return r
}
