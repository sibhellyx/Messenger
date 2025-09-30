package middleware

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sibhellyx/Messenger/internal/models/entity"
	"github.com/sibhellyx/Messenger/internal/models/payload"
)

type JwtManagerInterface interface {
	Parse(accessToken string) (payload.JwtPayload, error)
}

type SessionRepositoryInterface interface {
	GetSessionByUuid(uuid string) (*entity.Session, error)
	DeleteSessionByUuid(uuid string) error
}

func AuthMiddleware(m JwtManagerInterface, s SessionRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {

		token := c.GetHeader("Authorization")
		if !strings.HasPrefix(token, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid authorization header"})
			return
		}

		if token == "" {
			slog.Warn("missing authorization header")
			c.AbortWithStatusJSON(401, gin.H{"error": "Authorization header required"})
			return
		}

		token = strings.TrimPrefix(token, "Bearer ")
		payload, err := m.Parse(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		session, err := s.GetSessionByUuid(payload.Uuid)
		if err != nil || session == nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Session not found"})
			return
		}

		if time.Now().After(session.ExpiresAt) {
			s.DeleteSessionByUuid(payload.Uuid)
			c.AbortWithStatusJSON(401, gin.H{"error": "Session expired"})
			return
		}

		c.Set("uuid", payload.Uuid)
		c.Set("user_id", payload.UserId)

		c.Next()
	}
}
