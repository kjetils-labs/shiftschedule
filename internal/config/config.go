package config

import (
	"fmt"

	"github.com/shiftschedule/internal/helpers/path"
	"github.com/spf13/viper"
)

type Config struct {
	PostgresHostname  string `mapstructure:"POSTGRES_HOST"`
	PostgresPort      int    `mapstructure:"POSTGRES_PORT"`
	PostgresUsername  string `mapstructure:"POSTGRES_USER"`
	PostgresPassword  string `mapstructure:"POSTGRES_PASSWORD"`
	PostgresDatabase  string `mapstructure:"POSTGRES_DB"`
	PostgresEnableTLS bool   `mapstructure:"POSTGRES_ENABLE_TLS"`
}

func Init() (*Config, error) {

	envName := ".env"
	var config Config
	viper.SetConfigType("env")
	abspath, err := path.FindFile(envName)
	if err != nil {
		return nil, err
	}
	viper.SetConfigFile(abspath)
	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = viper.BindEnv(
		"POSTGRES_HOST",
		"POSTGRES_PORT",
		"POSTGRES_USER",
		"POSTGRES_PASSWORD",
		"POSTGRES_DB",
		"POSTGRES_ENABLE_TLS",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bind environment variables. %w", err)
	}

	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read in configuration. %w", err)
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal to configuration. %w", err)
	}

	return &config, nil

}
