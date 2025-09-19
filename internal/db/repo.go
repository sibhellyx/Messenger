package db

import (
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
