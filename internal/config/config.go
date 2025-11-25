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
	APIHostname       string `mapstructure:"API_HOST"`
	APIPort           int    `mapstructure:"API_PORT"`
}

func Init() (*Config, error) {

	envName := ".env"
	var config Config
	viper.SetDefault("POSTGRES_PORT", 5432)
	viper.SetDefault("POSTGRES_ENABLE_TLS", true)
	viper.SetDefault("API_PORT", 8080)
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
		"API_HOST",
		"API_PORT",
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

	if config.PostgresHostname == "" {
		return nil, fmt.Errorf("POSTGRES_HOST is empty or did not bind properly")
	}

	if config.PostgresDatabase == "" {
		return nil, fmt.Errorf("POSTGRES_DB is empty or did not bind properly")
	}

	if config.PostgresPassword == "" {
		return nil, fmt.Errorf("POSTGRES_PASSWORD is empty or did not bind properly")
	}

	if config.APIHostname == "" {
		return nil, fmt.Errorf("API_HOST is empty or did not bind properly")
	}

	return &config, nil

}
