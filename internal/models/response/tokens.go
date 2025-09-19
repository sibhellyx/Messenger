package response

import (
	"errors"
	"log/slog"
)

type Tokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func (t Tokens) Validate() error {
	slog.Debug("validating tokens")
	if t.AccessToken == "" {
		slog.Error("access is required")
		return errors.New("access is required")
	}
	if t.RefreshToken == "" {
		slog.Error("refresh_token is required")
		return errors.New("refresh_token is required")
	}
	slog.Debug("validating tokens completed")
	return nil
}
