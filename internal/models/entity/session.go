package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// session model for jwt auth
type Session struct {
	gorm.Model
	UserID       uint      `gorm:"not null;index" json:"user_id"`
	UUID         uuid.UUID `gorm:"type:uuid;not null;unique_index" json:"uuid"`
	RefreshToken string    `gorm:"size:255" json:"refresh_token"`
	ExpiresAt    time.Time `gorm:"type:timestamptz" json:"expires_at"`
	UserAgent    string    `gorm:"size:255" json:"user_agent"`
	LastIP       string    `gorm:"size:255" json:"last_ip"`
	User         User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// creating uuid for unique session
func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.UUID == uuid.Nil {
		s.UUID = uuid.New()
	}
	return nil
}

// check expired session
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// check valid of session
func (s *Session) IsValid() bool {
	return !s.IsExpired() && s.RefreshToken != ""
}
