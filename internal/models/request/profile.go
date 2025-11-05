package request

import (
	"errors"
	"time"
)

type ProfileRequest struct {
	Avatar      string     `gorm:"type:varchar(500)" json:"avatar,omitempty"`
	Bio         string     `gorm:"type:text" json:"bio,omitempty"`
	DateOfBirth *time.Time `gorm:"type:date" json:"date_of_birth,omitempty"`
}

func (pr *ProfileRequest) Validate() error {
	if pr.DateOfBirth != nil {
		if pr.DateOfBirth.After(time.Now()) {
			return errors.New("failed validate date_of_birth: cannot be in the future")
		}
	}

	return nil
}
