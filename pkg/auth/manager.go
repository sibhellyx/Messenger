package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/sibhellyx/Messenger/internal/models/payload"
)

type Manager struct {
	signingKey string
	logger     *slog.Logger
}

func NewManager(singingKey string) *Manager {
	return &Manager{signingKey: singingKey}
}

func (m *Manager) NewJWT(payload payload.JwtPayload, ttl time.Duration) (string, error) {
	m.logger.Debug("creating jwt")
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(ttl).Unix(),
		Subject:   payload.UserId,
		Id:        payload.Uuid,
	})
	return token.SignedString([]byte(m.signingKey))
}

func (m *Manager) Parse(accessToken string) (payload.JwtPayload, error) {
	m.logger.Debug("parse jwt")
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (i interface{}, err error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			errMsg := "unexpected signing method"
			m.logger.Error("jwt parsing failed",
				"error", errMsg,
				"algorithm", token.Header["alg"])
			return nil, errors.New(errMsg)
		}
		return []byte(m.signingKey), nil
	})
	if err != nil {
		m.logger.Error("jwt parsing failed", "error", err)
		return payload.JwtPayload{}, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		errMsg := "error get user from token"
		m.logger.Error("jwt parsing failed", "error", errMsg)
		return payload.JwtPayload{}, errors.New(errMsg)
	}

	payload := payload.JwtPayload{
		UserId: claims["sub"].(string),
		Uuid:   claims["jti"].(string),
	}

	m.logger.Debug("parsed jwt", "payload", payload)

	return payload, nil
}

func (m *Manager) NewRefreshToken() (string, error) {
	m.logger.Debug("creating refresh token")
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		m.logger.Error("failed to generate refresh token", "error", err)
		return "", err
	}

	return hex.EncodeToString(b), nil
}
