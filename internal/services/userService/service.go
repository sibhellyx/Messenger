package userservice

import (
	"log/slog"

	"github.com/sibhellyx/Messenger/internal/models/entity"
)

type UserRepositoryInterface interface {
	// get users
	GetUsers(search string) ([]*entity.User, error)
}

type UserService struct {
	repository UserRepositoryInterface
}

func NewUserService(repo UserRepositoryInterface) *UserService {
	return &UserService{
		repository: repo,
	}
}

func (s *UserService) GetUsers(search string) ([]*entity.User, error) {
	users, err := s.repository.GetUsers(search)
	if err != nil {
		slog.Error("failed to get users from repository", "error", err)
		return nil, err
	}

	for _, user := range users {
		user.Password = ""
	}

	return users, nil
}
