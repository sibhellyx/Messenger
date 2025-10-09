package request

import (
	"errors"
	"log/slog"
)

type ParticipantRequest struct {
	Id     string `json:"chat_id"`
	UserId string `json:"user_id"`
}

func (r ParticipantRequest) Validate() error {
	slog.Debug("validating participant input")
	if r.Id == "" {
		slog.Error("chat_id is required")
		return errors.New("chat_id is required")
	}
	if r.UserId == "" {
		slog.Error("user_id is required")
		return errors.New("user_id is required")
	}
	slog.Debug("validating participant request completed")
	return nil
}

type ParticipantUpdateRequest struct {
	Id     string `json:"chat_id"`
	UserId string `json:"user_id"`
	Role   string `json:"role"`
}

func (r ParticipantUpdateRequest) Validate() error {
	slog.Debug("validating participant input")
	if r.Id == "" {
		slog.Error("chat_id is required")
		return errors.New("chat_id is required")
	}
	if r.UserId == "" {
		slog.Error("user_id is required")
		return errors.New("user_id is required")
	}
	if r.Role == "" {
		slog.Error("role is required")
		return errors.New("role is required")
	}
	slog.Debug("validating participant request completed")
	return nil
}
