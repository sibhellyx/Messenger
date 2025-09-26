package payload

import (
	"errors"
	"log/slog"
)

type PayloadForRefresh struct {
	UserId       string `json:"user_id"`
	Uuid         string `json:"uuid"`
	RefreshToken string `json:"refreshToken"`
}

func (p PayloadForRefresh) Validate() error {
	slog.Debug("validating tokens")
	if p.UserId == "" {
		slog.Error("user_id is required")
		return errors.New("user_id is required")
	}
	if p.Uuid == "" {
		slog.Error("uuid is required")
		return errors.New("uuid is required")
	}
	if p.RefreshToken == "" {
		slog.Error("refresh_token is required")
		return errors.New("refresh_token is required")
	}
	slog.Debug("validating tokens completed")
	return nil
}
