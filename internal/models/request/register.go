package request

import (
	"errors"
	"log/slog"
)

type RegisterRequest struct {
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Tgname   string `json:"tg_username"`
	Password string `json:"password"`
}

func (r RegisterRequest) Validate() error {
	slog.Debug("validating user registration")
	if r.Name == "" {
		slog.Error("name is required")
		return errors.New("name is required")
	}
	if r.Surname == "" {
		slog.Error("surname is required")
		return errors.New("surname is required")
	}
	if r.Tgname == "" {
		slog.Error("tg_username is required")
		return errors.New("tg_username is required")
	}
	if r.Password == "" {
		slog.Error("password is required")
		return errors.New("password is required")
	}
	slog.Debug("validating user registration completed")
	return nil
}
