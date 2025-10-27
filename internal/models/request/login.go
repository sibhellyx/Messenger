package request

import (
	"errors"
	"log/slog"
)

type LoginRequest struct {
	Tgname   string `json:"tg_username"`
	Password string `json:"password"`
}

type LoginParams struct {
	UserAgent string
	LastIp    string
}

type VerifyCodeRequest struct {
	UserID string `json:"user_id"`
	Code   string `json:"code"`
}

func (l LoginRequest) Validate() error {
	slog.Debug("validating login input")
	if l.Tgname == "" {
		slog.Error("tg_username is required")
		return errors.New("tg_username is required")
	}
	if l.Password == "" {
		slog.Error("password is required")
		return errors.New("password is required")
	}
	slog.Debug("validating login input completed")
	return nil
}

func (r VerifyCodeRequest) Validate() error {
	slog.Debug("validating verify code request")
	if r.Code == "" {
		slog.Error("code is required")
		return errors.New("code is required")
	}
	if r.UserID == "" {
		slog.Error("user_id is required")
		return errors.New("user_id is required")
	}
	return nil
}
