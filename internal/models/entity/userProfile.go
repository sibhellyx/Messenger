package entity

import (
	"time"

	"gorm.io/gorm"
)

type UserProfile struct {
	gorm.Model
	Avatar      string     `gorm:"type:varchar(500)" json:"avatar"`
	Bio         string     `gorm:"type:text" json:"bio"`
	DateOfBirth *time.Time `gorm:"type:date" json:"date_of_birth,omitempty"`
	User        *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
