// entity/message.go
package entity

import (
	"errors"

	"gorm.io/gorm"
)

type MessageType string

const (
	MessageTypeText   MessageType = "text"
	MessageTypeImage  MessageType = "image"
	MessageTypeFile   MessageType = "file"
	MessageTypeSystem MessageType = "system"
)

type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
)

type Message struct {
	gorm.Model
	ChatID   uint          `gorm:"not null;index" json:"chatId"`
	UserID   uint          `gorm:"not null;index" json:"userId"`
	Type     MessageType   `gorm:"type:varchar(50);default:'text'" json:"type"`
	Content  string        `gorm:"type:text;not null" json:"content"`
	Status   MessageStatus `gorm:"type:varchar(50);default:'sent'" json:"status"`
	ClientID string        `gorm:"type:varchar(100);index" json:"clientId"`

	// for files and images
	FileURL  string `gorm:"type:varchar(500)" json:"fileUrl,omitempty"`
	FileName string `gorm:"type:varchar(255)" json:"fileName,omitempty"`
	FileSize int64  `json:"fileSize,omitempty"`
	MimeType string `gorm:"type:varchar(100)" json:"mimeType,omitempty"`

	// GORM relationships
	Chat      *Chat     `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Replies   []Message `gorm:"foreignKey:ReplyToID" json:"replies,omitempty"`
	ReplyToID uint      `gorm:"index" json:"replyToId,omitempty"`
	ReplyTo   *Message  `gorm:"foreignKey:ReplyToID" json:"replyTo,omitempty"`
}

func (Message) TableName() string {
	return "messages"
}

func (m Message) Validate() error {
	if m.ChatID == 0 {
		return errors.New("chatId is required")
	}

	if m.UserID == 0 {
		return errors.New("userId is required")
	}

	if m.Content == "" && m.Type == MessageTypeText {
		return errors.New("content is required for text messages")
	}

	if !m.isValidType() {
		return errors.New("invalid message type")
	}

	if !m.isValidStatus() {
		return errors.New("invalid message status")
	}

	return nil
}

func (m Message) isValidType() bool {
	switch m.Type {
	case MessageTypeText, MessageTypeImage, MessageTypeFile, MessageTypeSystem:
		return true
	default:
		return false
	}
}

func (m Message) isValidStatus() bool {
	switch m.Status {
	case MessageStatusSent, MessageStatusDelivered, MessageStatusRead:
		return true
	default:
		return false
	}
}
