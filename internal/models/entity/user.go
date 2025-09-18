package entity

import (
	"errors"
	"log/slog"

	"gorm.io/gorm"
)

// model of user
type User struct {
	gorm.Model
	Name     string    `gorm:"size:50;not null" json:"name"`
	Surname  string    `gorm:"size:50;not null" json:"surname"`
	Tgname   string    `gorm:"type:varchar(50);unique_index;not null" json:"tg_username"`
	Password string    `gorm:"size:255;not null" json:"password"`
	Sessions []Session `gorm:"foreignKey:UserID" json:"sessions,omitempty"`
}

func (User) TableName() string {
	return "users"
}

// validating input for registratiom
func (user User) Validate() error {
	slog.Debug("validating user registration")
	if user.Name == "" {
		slog.Error("name is required")
		return errors.New("name is required")
	}
	if user.Surname == "" {
		slog.Error("surname is required")
		return errors.New("surname is required")
	}
	if user.Tgname == "" {
		slog.Error("tg_username is required")
		return errors.New("tg_username is required")
	}
	if user.Password == "" {
		slog.Error("password is required")
		return errors.New("password is required")
	}
	slog.Debug("validating user registration completed")
	return nil
}
