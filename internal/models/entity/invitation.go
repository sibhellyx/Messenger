package entity

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type InvitationStatus string

const (
	InvitationPending  InvitationStatus = "pending"
	InvitationAccepted InvitationStatus = "accepted"
	InvitationDeclined InvitationStatus = "declined"
	InvitationExpired  InvitationStatus = "expired"
)

type ChatInvitation struct {
	gorm.Model
	ChatID        uint             `gorm:"not null;index" json:"chatId"`
	InvitedBy     uint             `gorm:"not null" json:"invitedBy"`
	InvitedUserID uint             `gorm:"index" json:"invitedUserId"`
	InvitedEmail  string           `gorm:"type:varchar(255);index" json:"invitedEmail"`
	Token         string           `gorm:"type:varchar(100);uniqueIndex;not null" json:"token"`
	Status        InvitationStatus `gorm:"type:varchar(50);default:'pending'" json:"status"`
	ExpiresAt     time.Time        `gorm:"not null" json:"expiresAt"`

	// GORM relationships
	Chat          *Chat `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
	InvitedByUser *User `gorm:"foreignKey:InvitedBy" json:"invitedByUser,omitempty"`
	InvitedUser   *User `gorm:"foreignKey:InvitedUserID" json:"invitedUser,omitempty"`
}

func (ChatInvitation) TableName() string {
	return "chat_invitations"
}

func (ci ChatInvitation) Validate() error {
	if ci.ChatID == 0 {
		return errors.New("chatId is required")
	}

	if ci.InvitedBy == 0 {
		return errors.New("invitedBy is required")
	}

	if ci.InvitedUserID == 0 && ci.InvitedEmail == "" {
		return errors.New("either invitedUserId or invitedEmail is required")
	}

	if ci.Token == "" {
		return errors.New("token is required")
	}

	if ci.ExpiresAt.IsZero() {
		return errors.New("expiresAt is required")
	}

	if !ci.isValidStatus() {
		return errors.New("invalid invitation status")
	}

	return nil
}

func (ci ChatInvitation) isValidStatus() bool {
	switch ci.Status {
	case InvitationPending, InvitationAccepted, InvitationDeclined, InvitationExpired:
		return true
	default:
		return false
	}
}

func (ci ChatInvitation) IsExpired() bool {
	return time.Now().After(ci.ExpiresAt)
}

func (ci ChatInvitation) CanBeAccepted() bool {
	return ci.Status == InvitationPending && !ci.IsExpired()
}
