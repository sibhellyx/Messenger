package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/spf13/viper"
)

type WsConfig struct {
	WriteWait      time.Duration `mapstructure:"WRITE_WAIT"`       // timeout for writing
	PongWait       time.Duration `mapstructure:"PONG_WAIT"`        // waiting pong messages
	PingPeriod     time.Duration `mapstructure:"PING_PERIOD"`      // period send ping
	MaxMessageSize int64         `mapstructure:"MAX_MESSAGE_SIZE"` // max size message
}

func LoadWsConfig() WsConfig {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	v := viper.New()
	v.SetConfigName("wsconfig")
	v.SetConfigType("yml")
	v.AddConfigPath("./configs")

	v.SetDefault("WRITE_WAIT", 10*time.Second)
	v.SetDefault("PONG_WAIT", 60*time.Second)
	v.SetDefault("PING_PERIOD", 54*time.Second)
	v.SetDefault("MAX_MESSAGE_SIZE", 512)

	var cfg WsConfig

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

	return cfg
}
