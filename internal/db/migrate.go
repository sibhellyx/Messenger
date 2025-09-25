package db

import (
	"log/slog"

	"github.com/sibhellyx/Messenger/internal/models/entity"
	"gorm.io/gorm"
)

// creating migration
func Migrate(db *gorm.DB) error {
	slog.Info("starting database migration")

	// Создаем расширение для UUID (если используете PostgreSQL)
	// if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
	// 	logger.Warn("failed to create uuid extension", "error", err)
	// }

	// complete auto migrate
	err := db.AutoMigrate(
		&entity.User{},
		&entity.Session{},
		&entity.Chat{},
		&entity.ChatParticipant{},
		&entity.Message{},
	)

	if err != nil {
		slog.Error("database migration failed", "error", err)
		return err
	}

	slog.Info("database migration completed successfully")
	return nil
}
