package request

import (
	"fmt"
	"time"
)

type ProfileRequest struct {
	Avatar      *string `json:"avatar,omitempty"`
	Bio         *string `json:"bio,omitempty"`
	DateOfBirth *string `json:"date_of_birth,omitempty"`
}

func (pr *ProfileRequest) Validate() error {
	if pr.DateOfBirth != nil && *pr.DateOfBirth != "" {
		_, err := time.Parse("2006-01-02", *pr.DateOfBirth)
		if err != nil {
			return fmt.Errorf("invalid date format: %w", err)
		}
	}
	return nil
}

func (pr *ProfileRequest) ParsedDate() (*time.Time, error) {
	if pr.DateOfBirth == nil || *pr.DateOfBirth == "" {
		return nil, nil
	}

	t, err := time.Parse("2006-01-02", *pr.DateOfBirth)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
