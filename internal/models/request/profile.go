package request

import (
	"errors"
	"fmt"
	"time"
)

type ProfileRequest struct {
	Avatar      string `json:"avatar,omitempty"`
	Bio         string `json:"bio,omitempty"`
	DateOfBirth string `json:"date_of_birth,omitempty" binding:"omitempty,datetime=2006-01-02"`
}

func (pr *ProfileRequest) ParsedDate() (*time.Time, error) {
	if pr.DateOfBirth == "" {
		return nil, nil
	}

	t, err := time.Parse("2006-01-02", pr.DateOfBirth)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (pr *ProfileRequest) Validate() error {
	if pr.DateOfBirth != "" {
		parsedDate, err := pr.ParsedDate()
		if err != nil {
			return fmt.Errorf("invalid date format: %w", err)
		}
		if parsedDate.After(time.Now()) {
			return errors.New("failed validate date_of_birth: cannot be in the future")
		}
	}
	return nil
}
