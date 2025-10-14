package config

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/viper"
)

type EnvConfig struct {
	Environment string `mapstructure:"ENVIRONMENT"` // type app
}

type ServerConfig struct {
	Port string `mapstructure:"PORT"`
}
type DbConfig struct {
	DBname     string `mapstructure:"DB_NAME"`
	DBuser     string `mapstructure:"DB_USER"`
	DBhost     string `mapstructure:"DB_HOST"`
	DBport     string `mapstructure:"DB_PORT"`
	DBpassword string `mapstructure:"DB_PASS"`
}

type AuthConfig struct {
	Salt       string `mapstructure:"SALT"`        // salt for hash
	SigningKey string `mapstructure:"SIGNING_KEY"` // signing key for auth manager
}

type JwtConfig struct {
	AccessTTL      int `mapstructure:"ACCESS_TTL"`
	RefreshTTL     int `mapstructure:"REFRESH_TTL"`
	ActiveSessions int `mapstructure:"ACTIVE_SESSIONS"`
}

type WsConfig struct {
	WriteWait      time.Duration `mapstructure:"WRITE_WAIT"`       // timeout for writing
	PongWait       time.Duration `mapstructure:"PONG_WAIT"`        // waiting pong messages
	PingPeriod     time.Duration `mapstructure:"PING_PERIOD"`      // period send ping
	MaxMessageSize int64         `mapstructure:"MAX_MESSAGE_SIZE"` // max size message
}

type Config struct {
	Env  EnvConfig
	Srv  ServerConfig
	Db   DbConfig
	Jwt  JwtConfig
	Auth AuthConfig
	Ws   WsConfig
}

func LoadConfig() (Config, error) {
	v := viper.New()
	v.AutomaticEnv()

	// Set default values for all configuration options
	setDefaults(v)

	var cfg Config

	// First, try to load YAML config
	v.SetConfigName("config")
	v.SetConfigType("yml")
	v.AddConfigPath("./configs")

	yamlErr := v.ReadInConfig()
	if yamlErr != nil {
		if _, ok := yamlErr.(viper.ConfigFileNotFoundError); ok {
			slog.Info("YAML config file not found, using defaults and environment variables")
		} else {
			slog.Warn("error reading YAML config file, using defaults and environment variables",
				slog.String("error", yamlErr.Error()))
		}
	} else {
		slog.Info("using YAML config file",
			slog.String("file", v.ConfigFileUsed()))
	}

	// Then, try to load .env file separately
	envViper := viper.New()
	envViper.SetConfigFile("configs/.env")
	envViper.SetConfigType("env")
	envViper.AutomaticEnv()

	envErr := envViper.ReadInConfig()
	if envErr != nil {
		if _, ok := envErr.(viper.ConfigFileNotFoundError); ok {
			slog.Info(".env file not found, using environment variables only")
		} else {
			slog.Warn("error reading .env file, using environment variables only",
				slog.String("error", envErr.Error()))
		}
	} else {
		slog.Info("using .env file",
			slog.String("file", envViper.ConfigFileUsed()))
		// Merge .env settings into main viper instance
		for _, key := range envViper.AllKeys() {
			v.Set(key, envViper.Get(key))
		}
	}

	if err := v.Unmarshal(&cfg.Env); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal env config: %w", err)
	}
	if err := v.Unmarshal(&cfg.Srv); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal server config: %w", err)
	}
	if err := v.Unmarshal(&cfg.Db); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal database config: %w", err)
	}
	if err := v.Unmarshal(&cfg.Auth); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal auth config: %w", err)
	}
	if err := v.Unmarshal(&cfg.Jwt); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal jwt config: %w", err)
	}
	if err := v.Unmarshal(&cfg.Ws); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal websocket config: %w", err)
	}

	// Log important configuration (without sensitive data)
	slog.Info("configuration loaded successfully",
		slog.String("port", cfg.Srv.Port),
		slog.String("db_user", cfg.Db.DBuser),
		slog.String("db_host", cfg.Db.DBhost),
		slog.String("db_name", cfg.Db.DBname),
		slog.String("environment", cfg.Env.Environment))

	slog.Info("auth configuration",
		"salt_length", len(cfg.Auth.Salt),
		"signing_key_length", len(cfg.Auth.SigningKey),
		"access_ttl", cfg.Jwt.AccessTTL,
		"refresh_ttl", cfg.Jwt.RefreshTTL,
		"active_sessions", cfg.Jwt.ActiveSessions)

	slog.Info("websocket configuration",
		"write_wait", cfg.Ws.WriteWait,
		"pong_wait", cfg.Ws.PongWait,
		"ping_period", cfg.Ws.PingPeriod,
		"max_message_size", cfg.Ws.MaxMessageSize)

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	// Environment defaults
	v.SetDefault("ENVIRONMENT", "development")

	// Server defaults
	v.SetDefault("PORT", "8080")

	// Database defaults
	v.SetDefault("DB_NAME", "db")
	v.SetDefault("DB_USER", "user")
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", "9920")
	v.SetDefault("DB_PASS", "password")

	// Authentication defaults
	v.SetDefault("SALT", "salt")
	v.SetDefault("SIGNING_KEY", "some_auth_key")

	// JWT defaults
	v.SetDefault("ACCESS_TTL", 15)
	v.SetDefault("REFRESH_TTL", 10)
	v.SetDefault("ACTIVE_SESSIONS", 5)

	// WebSocket defaults
	v.SetDefault("WRITE_WAIT", 10*time.Second)
	v.SetDefault("PONG_WAIT", 60*time.Second)
	v.SetDefault("PING_PERIOD", 54*time.Second)
	v.SetDefault("MAX_MESSAGE_SIZE", 512)
}

func (cfg Config) GetDbString() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", cfg.Db.DBhost, cfg.Db.DBuser, cfg.Db.DBpassword, cfg.Db.DBname, cfg.Db.DBport)
}
