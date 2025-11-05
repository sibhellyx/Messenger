package migrate

import (
	"log/slog"

	"github.com/sibhellyx/Messenger/internal/models/entity"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	slog.Info("starting database migration with existence check")

	migrationOrder := []struct {
		model     interface{}
		tableName string
	}{
		{&entity.User{}, "users"},
		{&entity.UserProfile{}, "profiles"},
		{&entity.Session{}, "sessions"},
		{&entity.Chat{}, "chats"},
		{&entity.Message{}, "messages"},
		{&entity.ChatParticipant{}, "chat_participants"},
	}

	for i, migration := range migrationOrder {
		slog.Info("migrating table", "order", i+1, "table", migration.tableName)

		if db.Migrator().HasTable(migration.model) {
			slog.Info("table already exists, updating schema", "table", migration.tableName)
		} else {
			slog.Info("creating new table", "table", migration.tableName)
		}

		if err := db.AutoMigrate(migration.model); err != nil {
			slog.Error("failed to migrate table",
				"table", migration.tableName,
				"order", i+1,
				"error", err)
			return err
		}

		slog.Info("successfully migrated table", "table", migration.tableName)
	}

	slog.Info("database migration completed successfully")
	return nil
}
