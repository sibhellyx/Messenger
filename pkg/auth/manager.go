package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sibhellyx/Messenger/internal/models/payload"
)

type Manager struct {
	signingKey []byte
}

func NewManager(signingKey string) *Manager {
	return &Manager{
		signingKey: []byte(signingKey),
	}
}

func (m *Manager) NewJWT(p payload.JwtPayload, ttl time.Duration) (string, error) {
	slog.Debug("creating jwt")

	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Subject:   p.UserId,
		ID:        p.Uuid,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	tokenString, err := token.SignedString(m.signingKey)
	if err != nil {
		slog.Error("failed to sign JWT token", "error", err)
		return "", err
	}

	slog.Debug("JWT token created successfully")
	return tokenString, nil
}

func (m *Manager) Parse(accessToken string) (payload.JwtPayload, error) {
	slog.Debug("parsing JWT token")

	token, err := jwt.ParseWithClaims(accessToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			errMsg := "unexpected signing method"
			slog.Error("JWT parsing failed",
				"error", errMsg,
				"algorithm", token.Method.Alg())
			return nil, errors.New(errMsg)
		}
		return m.signingKey, nil
	})

	if err != nil {
		slog.Error("JWT parsing failed", "error", err)
		return payload.JwtPayload{}, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		errMsg := "invalid token claims"
		slog.Error("JWT parsing failed", "error", errMsg)
		return payload.JwtPayload{}, errors.New(errMsg)
	}

	if time.Now().After(claims.ExpiresAt.Time) {
		errMsg := "token expired"
		slog.Error("JWT parsing failed", "error", errMsg)
		return payload.JwtPayload{}, errors.New(errMsg)
	}

	result := payload.JwtPayload{
		UserId: claims.Subject,
		Uuid:   claims.ID,
	}

	slog.Debug("JWT token parsed successfully",
		"user_id", result.UserId,
		"uuid", result.Uuid,
	)

	return result, nil
}

func (m *Manager) ParseIgnoreExpiration(accessToken string) (payload.JwtPayload, error) {
	slog.Debug("parsing JWT token (ignoring expiration)")

	token, err := jwt.ParseWithClaims(accessToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			errMsg := "unexpected signing method"
			slog.Error("JWT parsing failed",
				"error", errMsg,
				"algorithm", token.Method.Alg())
			return nil, errors.New(errMsg)
		}
		return m.signingKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			slog.Debug("JWT token expired, but parsing claims anyway")
		} else {
			slog.Error("JWT parsing failed", "error", err)
			return payload.JwtPayload{}, err
		}
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		errMsg := "invalid token claims"
		slog.Error("JWT parsing failed", "error", errMsg)
		return payload.JwtPayload{}, errors.New(errMsg)
	}

	result := payload.JwtPayload{
		UserId: claims.Subject,
		Uuid:   claims.ID,
	}

	slog.Debug("JWT token parsed successfully (ignoring expiration)",
		"user_id", result.UserId,
		"uuid", result.Uuid,
	)

	return result, nil
}

func (m *Manager) NewRefreshToken() (string, error) {
	slog.Debug("generating refresh token")

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		slog.Error("failed to generate refresh token", "error", err)
		return "", err
	}

	token := hex.EncodeToString(b)
	slog.Debug("refresh token generated successfully")
	return token, nil
}
