package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sibhellyx/Messenger/internal/middleware"
)

type AuthHandlerInterface interface {
	LogoutUser(c *gin.Context)
	RefreshToken(c *gin.Context)
	Register(c *gin.Context)
	SignIn(c *gin.Context)
}

type WsHandlerInterface interface {
	Connect(c *gin.Context)
}

type ChatHandlerInterface interface {
	CreateChat(c *gin.Context)
	UpdateChat(c *gin.Context)
	DeleteChat(c *gin.Context)
	GetUserChats(c *gin.Context)
	GetChats(c *gin.Context)
	FindChats(c *gin.Context)
	AddParticipant(c *gin.Context)
	GetChatParticipants(c *gin.Context)
}

func CreateRoutes(
	authHandler AuthHandlerInterface,
	chatHandler ChatHandlerInterface,
	wsHandler WsHandlerInterface,
	m middleware.JwtManagerInterface,
	repo middleware.SessionRepositoryInterface,
) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.LoggingMiddleware())

	// auth endpoints
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.SignIn)
	r.POST("/refresh", middleware.AuthMiddleware(m, repo), authHandler.RefreshToken)
	r.POST("/logout", middleware.AuthMiddleware(m, repo), authHandler.LogoutUser)

	// chat enpoints
	r.POST("/chat/create", middleware.AuthMiddleware(m, repo), chatHandler.CreateChat)
	r.DELETE("/chat", middleware.AuthMiddleware(m, repo), chatHandler.DeleteChat)
	r.PUT("/chat", middleware.AuthMiddleware(m, repo), chatHandler.UpdateChat)
	r.GET("/chats", middleware.AuthMiddleware(m, repo), chatHandler.GetUserChats)
	r.GET("/chats/all", middleware.AuthMiddleware(m, repo), chatHandler.GetChats)
	r.GET("/chats/search", middleware.AuthMiddleware(m, repo), chatHandler.FindChats)
	// participants endpoints
	r.POST("/chat/add", middleware.AuthMiddleware(m, repo), chatHandler.AddParticipant)
	r.GET("/chat/participants", middleware.AuthMiddleware(m, repo), chatHandler.GetChatParticipants)

	// ws handlers
	r.GET("/connect", middleware.AuthMiddleware(m, repo), wsHandler.Connect)
	return r
}
