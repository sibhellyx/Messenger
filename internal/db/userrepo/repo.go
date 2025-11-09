package userrepo

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/sibhellyx/Messenger/internal/models/entity"
	"github.com/sibhellyx/Messenger/internal/models/response"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) GetUsers(search string) ([]*entity.User, error) {
	var users []*entity.User

	query := r.db.Model(&entity.User{})

	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(surname) LIKE ? OR LOWER(tgname) LIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	result := query.Select("id, name, surname, tgname, created_at, updated_at").Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}

func (r *UserRepository) GetUsersWithProfiles(search string) ([]*response.UserWithProfile, error) {
	var userProfiles []*entity.UserProfile

	query := r.db.Preload("User").
		Joins("JOIN users ON user_profiles.user_id = users.id").
		Where("users.deleted_at IS NULL")

	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(users.name) LIKE ? OR LOWER(users.surname) LIKE ? OR LOWER(users.tgname) LIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	result := query.Find(&userProfiles)
	if result.Error != nil {
		slog.Error("failed to get users with profiles", "error", result.Error, "search", search)
		return nil, errors.New("failed to get users with profiles")
	}

	usersWithProfiles := make([]*response.UserWithProfile, 0, len(userProfiles))

	for _, profile := range userProfiles {
		userWithProfile := &response.UserWithProfile{
			UserID:      profile.User.ID,
			Name:        profile.User.Name,
			Surname:     profile.User.Surname,
			Tgname:      profile.User.Tgname,
			Avatar:      profile.Avatar,
			Bio:         profile.Bio,
			DateOfBirth: profile.DateOfBirth,
		}
		usersWithProfiles = append(usersWithProfiles, userWithProfile)
	}

	return usersWithProfiles, nil
}

func (r *UserRepository) UpdateProfile(profile entity.UserProfile) error {
	updates := map[string]interface{}{
		"avatar":        profile.Avatar,
		"bio":           profile.Bio,
		"date_of_birth": profile.DateOfBirth,
	}

	result := r.db.Model(&entity.UserProfile{}).
		Where("user_id = ?", profile.UserID).
		Updates(updates)

	if result.Error != nil {
		slog.Error("failed to update user profile", "error", result.Error, "user_id", profile.UserID)
		return errors.New("failed update user profile")
	}

	if result.RowsAffected == 0 {
		profile.CreatedAt = time.Now()
		profile.UpdatedAt = time.Now()
		result = r.db.Create(&profile)
		if result.Error != nil {
			slog.Error("failed to create user profile", "error", result.Error, "user_id", profile.UserID)
			return errors.New("failed to create user profile")
		}
		slog.Info("user profile created successfully", "user_id", profile.UserID)
		return nil
	}

	slog.Info("user profile updated successfully", "user_id", profile.UserID)
	return nil
}

func (r *UserRepository) GetUserById(userId uint) (*entity.User, error) {
	var user entity.User
	result := r.db.First(&user, userId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			slog.Warn("user not found or invalid credentials", "user_id", userId)
			return nil, errors.New("invalid user_id")
		}
		slog.Error("failed to get user by id", "error", result.Error, "user_id", userId)
		return nil, errors.New("failed get user by user_id")
	}

	return &user, nil
}

func (r *UserRepository) GetFullInfoAboutUser(userId uint) (*response.UserWithProfile, error) {
	var profile entity.UserProfile
	result := r.db.Preload("User").Where("user_id = ?", userId).First(&profile)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			user, err := r.GetUserById(userId)
			if err != nil {
				return nil, errors.New("failed get user by user_id")
			}
			return &response.UserWithProfile{
				UserID:      user.ID,
				Name:        user.Name,
				Surname:     user.Surname,
				Tgname:      user.Tgname,
				Avatar:      "",
				Bio:         "",
				DateOfBirth: nil,
			}, nil
		}
		slog.Error("failed to get user profile", "error", result.Error, "user_id", userId)
		return nil, errors.New("failed to get user profile")
	}

	return &response.UserWithProfile{
		UserID:      profile.User.ID,
		Name:        profile.User.Name,
		Surname:     profile.User.Surname,
		Tgname:      profile.User.Tgname,
		Avatar:      profile.Avatar,
		Bio:         profile.Bio,
		DateOfBirth: profile.DateOfBirth,
	}, nil
}
