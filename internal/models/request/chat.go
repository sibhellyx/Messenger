package request

import "github.com/sibhellyx/Messenger/internal/models/entity"

type CreateChatRequest struct {
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Type        entity.ChatType `json:"type,omitempty"`
	IsPrivate   bool            `json:"is_private,omitempty"`
}
