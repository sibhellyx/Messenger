package authrepo

import (
	"errors"
	"log/slog"
	"time"

	"github.com/sibhellyx/Messenger/internal/models/entity"
	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{
		db: db,
	}
}

func (r *AuthRepository) CreateUser(user entity.User) error {
	slog.Debug("creating user", "tgname", user.Tgname)

	result := r.db.Create(&user)
	if result.Error != nil {
		slog.Error("failed to create user", "error", result.Error, "tgname", user.Tgname)
		return errors.New("failed create user")
	}

	slog.Info("user created successfully", "user_id", user.ID, "email", user.Tgname)
	return nil
}

func (r *AuthRepository) GetUserByCredentails(tgname, password string) (*entity.User, error) {
	slog.Debug("get user by credentails", "tgname", tgname)

	var user entity.User
	result := r.db.Where("tgname = ? AND password = ?", tgname, password).First(&user) //find user

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			slog.Warn("user not found or invalid credentials", "tgname", tgname)
			return nil, errors.New("invalid credentials")
		}
		slog.Error("database error", "error", result.Error, "tgname", tgname)
		return nil, result.Error
	}

	slog.Info("user authenticated successfully", "user_id", user.ID, "tgname", tgname)
	return &user, nil
}

func (r *AuthRepository) GetUserByTgname(tgname string) (*entity.User, error) {
	var user entity.User
	result := r.db.Where("tgname = ?", tgname).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			slog.Warn("user not found or invalid credentials", "tgname", tgname)
			return nil, errors.New("invalid credentials")
		}
		slog.Error("database error", "error", result.Error, "tgname", tgname)
		return nil, result.Error
	}
	return &user, nil
}

func (r *AuthRepository) CreateSession(session entity.Session) error {
	slog.Debug("create session user", "user_id", session.UserID)

	result := r.db.Create(&session)
	if result.Error != nil {
		slog.Error("failed to create user session", "error", result.Error, "user_id", session.UserID)
		return result.Error
	}
	slog.Info("session created successfully", "uuid", session.UUID, "refreshToken", session.RefreshToken, "user_id", session.UserID)
	return nil
}

func (r *AuthRepository) DeleteSessionByUuid(uuid string) error {
	return r.db.Where("uuid = ?", uuid).Delete(&entity.Session{}).Error
}

func (r *AuthRepository) UpdateSession(session entity.Session) error {
	result := r.db.Save(&session)
	if result.Error != nil {
		slog.Error("failed to update session", "error", result.Error, "uuid", session.UUID, "user_id", session.UserID)
		return result.Error
	}
	slog.Info("session updated successfully", "uuid", session.UUID, "refreshToken", session.RefreshToken, "user_id", session.UserID)
	return nil
}

func (r *AuthRepository) FindJwtSessionByUuidAndRefreshToken(uuid, refreshToken string) (*entity.Session, error) {
	slog.Debug("get session", "uuid", uuid, "refreshToken", refreshToken)

	var session entity.Session
	result := r.db.Where("uuid = ? AND refresh_token = ?", uuid, refreshToken).First(&session) //find session

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			slog.Warn("session with this token not found", "uuid", uuid, "refreshToken", refreshToken)
			return nil, errors.New("invalid credentials")
		}
		slog.Error("database error", "error", result.Error, "uuid", uuid, "refreshToken", refreshToken)
		return nil, errors.New("database error")
	}

	slog.Info("session founded successfully", "uuid", uuid, "refreshToken", refreshToken, "user_id", session.UserID)
	return &session, nil
}

func (r *AuthRepository) GetSessionByUuid(uuid string) (*entity.Session, error) {
	slog.Debug("get session by uuid", "uuid", uuid)
	var session entity.Session
	result := r.db.Where("uuid = ?", uuid).First(&session)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			slog.Warn("session for this uuid not found", "uuid", uuid)
			return nil, errors.New("session not found")
		}
		slog.Error("database error", "error", result.Error, "uuid", uuid)
		return nil, errors.New("database error")
	}
	return &session, nil
}

func (r *AuthRepository) DeleteExpiredSessions(userId uint) error {
	slog.Debug("deleting expired sessions", "user_id", userId)
	return r.db.Where("user_id = ? AND expires_at < ?", userId, time.Now()).
		Delete(&entity.Session{}).Error
}

func (r *AuthRepository) DeleteOldestSession(userId uint) error {
	slog.Debug("deleted oldest session", "user_id", userId)
	var session entity.Session
	err := r.db.Where("user_id = ?", userId).
		Order("created_at ASC").
		First(&session).Error
	if err != nil {
		slog.Error("database error", "error", err.Error(), "user_id", userId)
		return err
	}
	return r.db.Delete(&session).Error
}

func (r *AuthRepository) CountActiveSessions(userId uint) (int64, error) {
	slog.Debug("get count active sessions", "user_id", userId)
	var count int64
	err := r.db.Model(&entity.Session{}).
		Where("user_id = ? AND expires_at > ?", userId, time.Now()).
		Count(&count).Error
	if err != nil {
		slog.Error("database error", "error", err.Error(), "user_id", userId)
	}
	return count, err
}

func (r *AuthRepository) ActivateUser(userId uint) error {
	slog.Debug("activating user start", "user_id", userId)
	result := r.db.Model(&entity.User{}).Where("id = ?", userId).Update("is_active", true)
	if result.Error != nil {
		slog.Error("failed to activate user", "error", result.Error, "user_id", userId)
		return errors.New("failed activate user")
	}
	if result.RowsAffected == 0 {
		slog.Warn("user not found for activation", "user_id", userId)
		return errors.New("user not found")
	}
	slog.Debug("activating user completed", "user_id", userId)
	return nil
}
