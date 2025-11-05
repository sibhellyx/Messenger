package userservice

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/sibhellyx/Messenger/internal/models/entity"
	"github.com/sibhellyx/Messenger/internal/models/request"
	"github.com/sibhellyx/Messenger/internal/models/response"
)

type UserRepositoryInterface interface {
	GetUsers(search string) ([]*entity.User, error)
	UpdateProfile(profile entity.UserProfile) error
	GetUserById(userId uint) (*entity.User, error)
	GetFullInfoAboutUser(userId uint) (*response.UserWithProfile, error)
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

func (s *UserService) UpdateProfile(userID string, req request.ProfileRequest) error {
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		slog.Error("failed parse user_id to uint", "user_id", userID)
		return errors.New("failed parse user_id")
	}

	err = req.Validate()
	if err != nil {
		slog.Error("failed validate req", "user_id", userID, "err", err)
		return err
	}

	user, err := s.repository.GetFullInfoAboutUser(uint(id))
	if err != nil || user == nil {
		slog.Error("failed found user", "user_id", user.UserID)
		return err
	}

	if user.Bio != req.Bio && req.Bio != "" {
		user.Bio = req.Bio
	}
	if user.Avatar != req.Avatar && req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if user.DateOfBirth != req.DateOfBirth && req.DateOfBirth != nil {
		user.DateOfBirth = req.DateOfBirth
	}
	profile := entity.UserProfile{
		UserID:      uint(id),
		Bio:         req.Bio,
		Avatar:      req.Avatar,
		DateOfBirth: req.DateOfBirth,
	}

	return s.repository.UpdateProfile(profile)
}

func (s *UserService) GetFullInfoAboutUser(userID string) (*response.UserWithProfile, error) {
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		slog.Error("failed parse user_id to uint", "user_id", userID)
		return nil, errors.New("failed parse user_id")
	}

	return s.repository.GetFullInfoAboutUser(uint(id))
}
