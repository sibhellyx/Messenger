package authservice

import (
	"github.com/sibhellyx/Messenger/internal/db"
	"github.com/sibhellyx/Messenger/internal/models/entity"
)

type AuthService struct {
	repository *db.Repository
}

func NewAuthService(repository *db.Repository) *AuthService {
	return &AuthService{
		repository: repository,
	}
}

func (s *AuthService) RegisterUser(user entity.User) error {
	err := user.Validate()
	return err
}
