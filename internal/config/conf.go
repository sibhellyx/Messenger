package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	// serv cfg
	Port string `mapstructure:"PORT"`
	//db cfg
	DBname     string `mapstructure:"DB_NAME"`
	DBuser     string `mapstructure:"DB_USER"`
	DBhost     string `mapstructure:"DB_HOST"`
	DBport     string `mapstructure:"DB_PORT"`
	DBpassword string `mapstructure:"DB_PASS"`
	//for tokens
	AccessTTL  int `mapstructure:"ACCESS_TTL"`
	RefreshTTL int `mapstructure:"REFRESH_TTL"`
	//for sessions
	ActiveSessions int `mapstructure:"ACTIVE_SESSIONS"`
}

func LoadConfig() Config {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yml")
	v.AddConfigPath("./configs")

	// default values
	v.SetDefault("PORT", "8080")
	v.SetDefault("DB_NAME", "db")
	v.SetDefault("DB_USER", "user")
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", "9920")
	v.SetDefault("DB_PASS", "password")
	v.SetDefault("ACCESS_TTL", 15)
	v.SetDefault("REFRESH_TTL", 10)
	v.SetDefault("ACTIVE_SESSIONS", 5)

	var cfg Config

	v.AutomaticEnv()

	// reading cfg file
	err := v.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Info("config file not found, using defaults and environment variables")
		} else {
			logger.Warn("error reading config file, using defaults and environment variables",
				slog.String("error", err.Error()))
		}
	} else {
		logger.Info("using config file",
			slog.String("file", v.ConfigFileUsed()))
	}

	// unmarshal cfg
	err = v.Unmarshal(&cfg)
	if err != nil {
		logger.Error("failed to unmarshal config",
			slog.String("error", err.Error()))
		panic(err)
	}

	logger.Info("configuration loaded successfully",
		slog.String("port", cfg.Port),
		slog.String("db_host", cfg.DBhost),
		slog.String("db_name", cfg.DBname))

	return cfg
}

func (cfg Config) GetDbString() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", cfg.DBhost, cfg.DBuser, cfg.DBpassword, cfg.DBname, cfg.DBport)
}
