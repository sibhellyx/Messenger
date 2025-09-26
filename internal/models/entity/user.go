package entity

import (
	"gorm.io/gorm"
)

// model of user
type User struct {
	gorm.Model
	Name     string    `gorm:"size:50;not null" json:"name"`
	Surname  string    `gorm:"size:50;not null" json:"surname"`
	Tgname   string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"tgUsername"`
	Password string    `gorm:"size:255;not null" json:"-"`
	Sessions []Session `gorm:"foreignKey:UserID" json:"sessions,omitempty"`
}

func (User) TableName() string {
	return "users"
}
