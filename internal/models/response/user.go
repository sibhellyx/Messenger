package response

import "time"

type UserWithProfile struct {
	UserID      uint       `gorm:"not null;index" json:"user_id"`
	Name        string     `gorm:"size:50;not null" json:"name"`
	Surname     string     `gorm:"size:50;not null" json:"surname"`
	Tgname      string     `gorm:"type:varchar(50);uniqueIndex;not null" json:"tgUsername"`
	Avatar      string     `gorm:"type:varchar(500)" json:"avatar,omitempty"`
	Bio         string     `gorm:"type:text" json:"bio,omitempty"`
	DateOfBirth *time.Time `gorm:"type:date" json:"date_of_birth,omitempty"`
}
