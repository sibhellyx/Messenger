package db

import (
	"errors"
	"log/slog"

	"github.com/sibhellyx/Messenger/internal/models/entity"
	"gorm.io/gorm"
)

type Repository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewRepository(db *gorm.DB, logger *slog.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (r *Repository) CreateUser(user entity.User) error {
	r.logger.Debug("creating user", "tgname", user.Tgname)

	result := r.db.Create(&user)
	if result.Error != nil {
		r.logger.Error("failed to create user", "error", result.Error, "tgname", user.Tgname)
		return result.Error
	}

	r.logger.Info("user created successfully", "user_id", user.ID, "email", user.Tgname)
	return nil
}

func (r *Repository) GetUserByCredentails(tgname, password string) (*entity.User, error) {
	r.logger.Debug("get user by credentails", "tgname", tgname)

	var user entity.User
	result := r.db.Where("tgname = ? AND password = ?", tgname, password).First(&user) //find user

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.Warn("user not found or invalid credentials", "tgname", tgname)
			return nil, errors.New("invalid credentials")
		}
		r.logger.Error("database error", "error", result.Error, "tgname", tgname)
		return nil, result.Error
	}

	r.logger.Info("user authenticated successfully", "user_id", user.ID, "tgname", tgname)
	return &user, nil
}

func (r *Repository) CreateSession(session entity.Session) error {
	r.logger.Debug("create session user", "user_id", session.UserID)

	result := r.db.Create(&session)
	if result.Error != nil {
		r.logger.Error("failed to create user session", "error", result.Error, "user_id", session.UserID)
		return result.Error
	}
	r.logger.Info("session created successfully", "uuid", session.UUID, "refreshToken", session.RefreshToken, "user_id", session.UserID)
	return nil
}

func (r *Repository) DeleteSessionByUuid(uuid string) error {
	return r.db.Where("uuid = ?", uuid).Delete(&entity.Session{}).Error
}

func (r *Repository) UpdateSession(session entity.Session) error {
	result := r.db.Save(&session)
	if result.Error != nil {
		r.logger.Error("failed to update session", "error", result.Error, "uuid", session.UUID, "user_id", session.UserID)
		return result.Error
	}
	r.logger.Info("session updated successfully", "uuid", session.UUID, "refreshToken", session.RefreshToken, "user_id", session.UserID)
	return nil
}

func (r *Repository) FindJwtSessionByUuidAndRefreshToken(uuid, refreshToken string) (*entity.Session, error) {
	r.logger.Debug("get session", "uuid", uuid, "refreshToken", refreshToken)

	var session entity.Session
	result := r.db.Where("uuid = ? AND refresh_token = ?", uuid, refreshToken).First(&session) //find session

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.Warn("session with this token not found", "uuid", uuid, "refreshToken", refreshToken)
			return nil, errors.New("invalid credentials")
		}
		r.logger.Error("database error", "error", result.Error, "uuid", uuid, "refreshToken", refreshToken)
		return nil, result.Error
	}

	r.logger.Info("session founded successfully", "uuid", uuid, "refreshToken", refreshToken, "user_id", session.UserID)
	return &session, nil
}

// func (r *Repository) DeleteExpiredSessions(userID uint) error {
// 	return r.db.Where("user_id = ? AND expires_at < ?", userID, time.Now()).
// 		Delete(&entity.Session{}).Error
// }

// func (r *Repository) CountActiveSessions(userID uint) (int64, error) {
// 	var count int64
// 	err := r.db.Model(&entity.Session{}).
// 		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
// 		Count(&count).Error
// 	return count, err
// }

// func (r *Repository) DeleteOldestSessions(userID uint, count int) error {
// 	var sessions []entity.Session
// 	err := r.db.Where("user_id = ? AND expires_at > ?", userID, time.Now()).
// 		Order("created_at ASC").
// 		Limit(count).
// 		Find(&sessions).Error
// 	if err != nil {
// 		return err
// 	}

// 	if len(sessions) == 0 {
// 		return nil
// 	}

// 	var uuids []uuid.UUID
// 	for _, s := range sessions {
// 		uuids = append(uuids, s.UUID)
// 	}

// 	return r.db.Where("uuid IN ?", uuids).Delete(&entity.Session{}).Error
// }
