package entity

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type ParticipantRole string

const (
	RoleOwner  ParticipantRole = "owner"
	RoleAdmin  ParticipantRole = "admin"
	RoleMember ParticipantRole = "member"
)

type ChatParticipant struct {
	gorm.Model
	ChatID               uint            `gorm:"not null;index" json:"chatId"`
	UserID               uint            `gorm:"not null;index" json:"userId"`
	Role                 ParticipantRole `gorm:"type:varchar(50);default:'member'" json:"role"`
	JoinedAt             *time.Time      `gorm:"default:now()" json:"joinedAt"`
	LastReadMessageID    *uint           `gorm:"index" json:"lastReadMessageId,omitempty"`
	IsMuted              bool            `gorm:"default:false" json:"isMuted"`
	NotificationsEnabled bool            `gorm:"default:true" json:"notificationsEnabled"`

	// GORM relationships
	Chat            *Chat    `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
	User            *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	LastReadMessage *Message `gorm:"foreignKey:LastReadMessageID" json:"lastReadMessage,omitempty"`
}

func (ChatParticipant) TableName() string {
	return "chat_participants"
}

func GetRoleForUpdate(s string) (ParticipantRole, error) {
	switch s {
	case "admin":
		return RoleAdmin, nil
	case "member":
		return RoleMember, nil
	default:
		return "", errors.New("incorrect role for update")
	}
}
