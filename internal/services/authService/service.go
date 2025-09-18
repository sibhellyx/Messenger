package authservice

import (
	"github.com/sibhellyx/Messenger/internal/db"
)

type AuthService struct {
	repository *db.Repository
}

func NewAuthService(repository *db.Repository) *AuthService {
	return &AuthService{
		repository: repository,
	}
}
