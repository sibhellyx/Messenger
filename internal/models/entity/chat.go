package entity

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type ChatType string

const (
	ChatTypeDirect  ChatType = "direct"
	ChatTypeGroup   ChatType = "group"
	ChatTypeChannel ChatType = "channel"
)

type Chat struct {
	gorm.Model
	Name           string    `gorm:"size:255;not null" json:"name"`
	Description    string    `gorm:"type:text" json:"description"`
	Type           ChatType  `gorm:"type:varchar(50);not null;default:'group'" json:"type"`
	AvatarURL      string    `gorm:"type:varchar(500)" json:"avatarUrl"`
	CreatedBy      uint      `gorm:"not null" json:"createdBy"` // Foreign key to User ID
	IsPrivate      bool      `gorm:"default:false" json:"isPrivate"`
	MaxMembers     int       `gorm:"default:100" json:"maxMembers"`
	LastMessageID  uint      `gorm:"index" json:"lastMessageId"`
	LastActivityAt time.Time `gorm:"default:now()" json:"lastActivityAt"`

	// GORM relationships
	Creator      *User             `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Participants []ChatParticipant `gorm:"foreignKey:ChatID" json:"participants,omitempty"`
	LastMessage  *Message          `gorm:"foreignKey:LastMessageID" json:"lastMessage,omitempty"`
	Messages     []Message         `gorm:"foreignKey:ChatID" json:"messages,omitempty"`
}

func (Chat) TableName() string {
	return "chats"
}

func (chat Chat) Validate() error {
	if chat.Name == "" && chat.Type != ChatTypeDirect {
		return errors.New("chat name is required for group and channel types")
	}

	if chat.Type == "" {
		return errors.New("chat type is required")
	}

	if !chat.isValidChatType() {
		return errors.New("invalid chat type")
	}

	if chat.CreatedBy == 0 {
		return errors.New("createdBy is required")
	}

	return nil
}

func (chat Chat) isValidChatType() bool {
	switch chat.Type {
	case ChatTypeDirect, ChatTypeGroup, ChatTypeChannel:
		return true
	default:
		return false
	}
}
