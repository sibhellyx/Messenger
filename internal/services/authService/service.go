package authservice

import (
	"log/slog"

	"github.com/sibhellyx/Messenger/internal/db"
	"github.com/sibhellyx/Messenger/internal/models/entity"
	"github.com/sibhellyx/Messenger/pkg/hash"
)

type AuthService struct {
	repository *db.Repository
	logger     *slog.Logger
	hasher     *hash.Hasher
}

func NewAuthService(repository *db.Repository, logger *slog.Logger, hasher *hash.Hasher) *AuthService {
	return &AuthService{
		repository: repository,
		logger:     logger,
		hasher:     hasher,
	}
}

// register user service layer
func (s *AuthService) RegisterUser(user entity.User) error {
	s.logger.Debug("service register started")
	err := user.Validate() //function return error if user not isValid
	if err != nil {
		s.logger.Error("error validating user input", "error", err.Error())
		return err
	}
	user.Password, err = s.hasher.Hash(user.Password) //hash password
	if err != nil {
		s.logger.Error("failed hash password", "error", err.Error())
		return err
	}
	err = s.repository.CreateUser(user) //write to repo
	if err != nil {
		s.logger.Error("failed create user in repo", "error", err.Error())
		return err
	}
	s.logger.Debug("service register completed")
	return nil
}
