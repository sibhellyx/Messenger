package userservice

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/sibhellyx/Messenger/internal/models/entity"
	"github.com/sibhellyx/Messenger/internal/models/request"
	"github.com/sibhellyx/Messenger/internal/models/response"
)

type UserRepositoryInterface interface {
	GetUsers(search string) ([]*entity.User, error)
	GetUsersWithProfiles(search string) ([]*response.UserWithProfile, error)
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

func (s *UserService) GetUsersWithProfiles(search string) ([]*response.UserWithProfile, error) {
	users, err := s.repository.GetUsersWithProfiles(search)
	if err != nil {
		slog.Error("failed to get users from repository", "error", err)
		return nil, err
	}

	return users, nil
}

func (s *UserService) UpdateProfile(userID string, req request.ProfileRequest) error {
	if userID == "" {
		slog.Error("empty user_id provided")
		return errors.New("user_id is required")
	}

	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		slog.Error("failed parse user_id to uint", "user_id", userID, "err", err)
		return errors.New("invalid user_id format")
	}

	err = req.Validate()
	if err != nil {
		slog.Error("failed validate req", "user_id", userID, "err", err)
		return err
	}

	user, err := s.repository.GetFullInfoAboutUser(uint(id))
	if err != nil {
		slog.Error("failed to get user", "user_id", id, "err", err)
		return errors.New("user not found")
	}
	if user == nil {
		slog.Error("user not found", "user_id", id)
		return errors.New("user not found")
	}

	updated := false

	if req.Bio != nil {
		if *req.Bio != user.Bio {
			user.Bio = *req.Bio
			updated = true
			slog.Debug("bio updated", "user_id", user.UserID, "new_bio", *req.Bio)
		}
	}

	if req.Avatar != nil {
		if *req.Avatar != user.Avatar {
			user.Avatar = *req.Avatar
			updated = true
			slog.Debug("avatar updated", "user_id", user.UserID, "new_avatar", *req.Avatar)
		}
	}

	if req.DateOfBirth != nil {
		if *req.DateOfBirth == "" {
			if user.DateOfBirth != nil {
				user.DateOfBirth = nil
				updated = true
				slog.Debug("date_of_birth cleared", "user_id", user.UserID)
			}
		} else {
			data, err := req.ParsedDate()
			if err != nil {
				slog.Error("failed to parse date", "user_id", user.UserID, "err", err)
				return errors.New("invalid date format")
			}

			currentDate := user.DateOfBirth
			if !compareDates(currentDate, data) {
				user.DateOfBirth = data
				updated = true
				slog.Debug("date_of_birth updated", "user_id", user.UserID, "new_date", data)
			}
		}
	}

	if !updated {
		slog.Info("no changes detected", "user_id", user.UserID)
		return nil
	}

	profile := entity.UserProfile{
		UserID:      uint(id),
		Bio:         user.Bio,
		Avatar:      user.Avatar,
		DateOfBirth: user.DateOfBirth,
	}

	err = s.repository.UpdateProfile(profile)
	if err != nil {
		slog.Error("failed to update profile", "user_id", user.UserID, "err", err)
		return errors.New("failed to update profile")
	}

	slog.Info("profile updated successfully", "user_id", user.UserID)
	return nil
}

func compareDates(date1, date2 *time.Time) bool {
	if date1 == nil && date2 == nil {
		return true
	}
	if (date1 == nil && date2 != nil) || (date1 != nil && date2 == nil) {
		return false
	}
	return date1.Equal(*date2)
}

func (s *UserService) GetFullInfoAboutUser(userID string) (*response.UserWithProfile, error) {
	if userID == "" {
		slog.Error("user_id is required")
		return nil, errors.New("user_id can't be empty")
	}
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		slog.Error("failed parse user_id to uint", "user_id", userID)
		return nil, errors.New("failed parse user_id")
	}

	return s.repository.GetFullInfoAboutUser(uint(id))
}
