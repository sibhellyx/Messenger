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
	JoinedAt             time.Time       `gorm:"default:now()" json:"joinedAt"`
	InvitedBy            uint            `gorm:"index" json:"invitedBy"`
	LastReadMessageID    uint            `gorm:"index" json:"lastReadMessageId"`
	IsMuted              bool            `gorm:"default:false" json:"isMuted"`
	NotificationsEnabled bool            `gorm:"default:true" json:"notificationsEnabled"`

	// GORM relationships
	Chat            *Chat    `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
	User            *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Inviter         *User    `gorm:"foreignKey:InvitedBy" json:"inviter,omitempty"`
	LastReadMessage *Message `gorm:"foreignKey:LastReadMessageID" json:"lastReadMessage,omitempty"`
}

func (ChatParticipant) TableName() string {
	return "chat_participants"
}

func (cp ChatParticipant) Validate() error {
	if cp.ChatID == 0 {
		return errors.New("chatId is required")
	}

	if cp.UserID == 0 {
		return errors.New("userId is required")
	}

	if !cp.isValidRole() {
		return errors.New("invalid participant role")
	}

	return nil
}

func (cp ChatParticipant) isValidRole() bool {
	switch cp.Role {
	case RoleOwner, RoleAdmin, RoleMember:
		return true
	default:
		return false
	}
}

func (cp ChatParticipant) CanModifyChat() bool {
	return cp.Role == RoleOwner || cp.Role == RoleAdmin
}

func (cp ChatParticipant) CanInviteUsers() bool {
	return cp.Role == RoleOwner || cp.Role == RoleAdmin
}
