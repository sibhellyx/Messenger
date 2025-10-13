package config

import (
	"log/slog"

	"github.com/spf13/viper"
)

type AuthConfig struct {
	// salt for hash
	Salt string `mapstructure:"SALT"`
	// signing key for auth manager
	SigningKey string `mapstructure:"SIGNING_KEY"`
}

func LoadEnvAuthConfig() AuthConfig {

	viper.SetConfigFile("configs/.env")
	viper.AutomaticEnv()

	viper.SetDefault("SALT", "salt")
	viper.SetDefault("SIGNING_KEY", "some_auth_key")

	var config AuthConfig
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			slog.Info("config file not found, using defaults and environment variables for auth and hash")
		} else {
			slog.Warn("error reading config file, using defaults and environment variables for auth and hash",
				slog.String("error", err.Error()))
		}
	} else {
		slog.Info("using config file",
			slog.String("file", viper.ConfigFileUsed()))
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		slog.Error("failed to unmarshal config auth",
			slog.String("error", err.Error()))
		panic(err)
	}

	slog.Info("auth confs",
		"salt", config.Salt,
		"key", config.SigningKey,
	)
	return config
}
