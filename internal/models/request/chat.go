package request

import (
	"errors"
	"log/slog"

	"github.com/sibhellyx/Messenger/internal/models/entity"
)

type CreateChatRequest struct {
	Name         string          `json:"name,omitempty"`
	Description  *string         `json:"description,omitempty"`
	Type         entity.ChatType `json:"type,omitempty"`
	AvatarURL    *string         `json:"avatarUrl,omitempty"`
	IsPrivate    bool            `json:"is_private,omitempty"`
	Participants []Participant   `json:"participants,omitempty"`
}

type UpdateChatRequest struct {
	Id          string  `json:"chat_id"`
	Name        string  `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	AvatarURL   *string `json:"avatarUrl,omitempty"`
	IsPrivate   bool    `json:"is_private,omitempty"`
}

func (r UpdateChatRequest) Validate() error {
	if r.Id == "" {
		slog.Error("chat_id is required")
		return errors.New("chat_id is required")
	}
	return nil
}

type ChatRequest struct {
	Id string `json:"chat_id"`
}

func (r ChatRequest) Validate() error {
	slog.Debug("validating chat request")
	if r.Id == "" {
		slog.Error("id is required")
		return errors.New("id is required")
	}
	slog.Debug("validating chat request completed")
	return nil
}

type Participant struct {
	ID string `json:"id"`
}

func (r CreateChatRequest) Validate() error {
	slog.Debug("validating creating chat input")
	// if directed chat
	if r.Type == entity.ChatTypeDirect {
		if len(r.Participants) != 1 {
			slog.Error("directed chat is require only 1 participant")
			return errors.New("directed chat is require only 1 participant")
		}
		if r.Description != nil {
			slog.Error("directed chat сannot have descriptions")
			return errors.New("directed chat сannot have descriptions")
		}
		if r.Name != "" {
			slog.Error("directed chat cannot have name")
			return errors.New("directed chat cannot have name")
		}
		if r.AvatarURL != nil {
			slog.Error("directed chat сannot have avatar")
			return errors.New("directed chat сannot have descriptions")
		}
		return nil
	}

	// if group or channel
	if r.Name == "" {
		slog.Error("name is required")
		return errors.New("name chat is required")
	}
	// add validate url

	slog.Debug("validating creating caht input completed")
	return nil
}
