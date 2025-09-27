package request

import (
	"errors"
	"log/slog"

	"github.com/sibhellyx/Messenger/internal/models/entity"
)

type CreateChatRequest struct {
	Name         string          `json:"name"`
	Description  *string         `json:"description,omitempty"`
	Type         entity.ChatType `json:"type,omitempty"`
	IsPrivate    bool            `json:"is_private,omitempty"`
	Participants []Participant   `json:"participants,omitempty"`
}

type Participant struct {
	ID uint `json:"id"`
}

func (r CreateChatRequest) Validate() error {
	slog.Debug("validating creating chat input")
	if r.Name == "" {
		slog.Error("name is required")
		return errors.New("name chat is required")
	}
	if r.Type == entity.ChatTypeDirect && len(r.Participants) != 1 {
		slog.Error("directed chat is require only 1 participant")
		return errors.New("directed chat is require only 1 participant")
	}
	slog.Debug("validating creating caht input completed")
	return nil
}
