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
	VerifyLogin(c *gin.Context)
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
	LeaveChat(c *gin.Context)
	RemoveParticipant(c *gin.Context)
	UpdateParticipant(c *gin.Context)
}

type UserHandlerInterface interface {
	GetUsers(c *gin.Context)
	GetUsersWithProfiles(c *gin.Context)
	UpdateUserProfile(c *gin.Context)
	GetUserProfile(c *gin.Context)
	GetMyProfile(c *gin.Context)
}

type MessageHandlerInterface interface {
	SendMessage(c *gin.Context)
	GetMessages(c *gin.Context)
}

func CreateRoutes(
	authHandler AuthHandlerInterface,
	chatHandler ChatHandlerInterface,
	wsHandler WsHandlerInterface,
	messageHandler MessageHandlerInterface,
	userHandler UserHandlerInterface,
	m middleware.JwtManagerInterface,
	repo middleware.SessionRepositoryInterface,
) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.LoggingMiddleware())

	// auth endpoints
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.SignIn)
	r.POST("/login/verify", authHandler.VerifyLogin)
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
	r.POST("/chat/leave", middleware.AuthMiddleware(m, repo), chatHandler.LeaveChat)
	r.DELETE("/chat/remove", middleware.AuthMiddleware(m, repo), chatHandler.RemoveParticipant)
	r.PUT("/chat/participant", middleware.AuthMiddleware(m, repo), chatHandler.UpdateParticipant)

	// message sender handler
	r.POST("/message/send", middleware.AuthMiddleware(m, repo), messageHandler.SendMessage)
	r.GET("/chat/messages", middleware.AuthMiddleware(m, repo), messageHandler.GetMessages)

	// users
	r.GET("/users", middleware.AuthMiddleware(m, repo), userHandler.GetUsers)
	r.GET("/users/full", middleware.AuthMiddleware(m, repo), userHandler.GetUsersWithProfiles)
	r.GET("/my", middleware.AuthMiddleware(m, repo), userHandler.GetMyProfile)
	r.GET("/profile", middleware.AuthMiddleware(m, repo), userHandler.GetUserProfile)
	r.PUT("/profile", middleware.AuthMiddleware(m, repo), userHandler.UpdateUserProfile)
	// ws handlers
	r.GET("/connect", middleware.AuthMiddleware(m, repo), wsHandler.Connect)
	return r
}
