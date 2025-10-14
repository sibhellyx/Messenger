package request

import "github.com/sibhellyx/Messenger/internal/models/entity"

type CreateMessage struct {
	ChatID    string             `json:"chatId" binding:"required"`
	Content   string             `json:"content" binding:"required"`
	Type      entity.MessageType `json:"type" binding:"required,oneof=text image file system"`
	ReplyToID *uint              `json:"replyToId,omitempty"`
	FileURL   *string            `json:"fileUrl,omitempty"`
	FileName  *string            `json:"fileName,omitempty"`
	FileSize  *int64             `json:"fileSize,omitempty"`
	MimeType  *string            `json:"mimeType,omitempty"`
	ClientID  string             `json:"clientId" binding:"required"`
}
