package response

import "time"

type UserWithProfile struct {
	UserID      uint       `json:"user_id"`
	Name        string     `json:"name"`
	Surname     string     `json:"surname"`
	Tgname      string     `json:"tgUsername"`
	Avatar      string     `json:"avatar"`
	Bio         string     `json:"bio"`
	DateOfBirth *time.Time `json:"date_of_birth"`
}
