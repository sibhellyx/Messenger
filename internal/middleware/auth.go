package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sibhellyx/Messenger/internal/models/payload"
)

type JwtManagerInterface interface {
	Parse(accessToken string) (payload.JwtPayload, error)
}

type SessionRepositoryInterface interface {
	CheckSessionByUuid(uuid string) (bool, error)
}

func AuthMiddleware(m JwtManagerInterface, s SessionRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Auth")
		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		payload, err := m.Parse(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		exists, err := s.CheckSessionByUuid(payload.Uuid)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		if !exists {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		c.Set("uuid", payload.Uuid)
		c.Set("user_id", payload.UserId)

		c.Next()
	}
}
