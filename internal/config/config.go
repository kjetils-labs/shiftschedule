// Package config loads the application configuration from environment variables, optionally seeded from a .env file.
package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/shiftschedule/internal/helpers/path"
	"github.com/spf13/viper"
)

const (
	AuthModeDev  = "dev"
	AuthModeOIDC = "oidc"
)

type Config struct {
	PostgresHostname  string `mapstructure:"POSTGRES_HOST"`
	PostgresUsername  string `mapstructure:"POSTGRES_USER"`
	PostgresPassword  string `mapstructure:"POSTGRES_PASSWORD"`
	PostgresDatabase  string `mapstructure:"POSTGRES_DB"`
	APIHostname       string `mapstructure:"API_HOST"`
	AuthMode          string `mapstructure:"AUTH_MODE"`
	DevUserEmail      string `mapstructure:"DEV_USER_EMAIL"`
	OIDCIssuerURL     string `mapstructure:"OIDC_ISSUER_URL"`
	OIDCClientID      string `mapstructure:"OIDC_CLIENT_ID"`
	Timezone          string `mapstructure:"TIMEZONE"`
	PostgresPort      int    `mapstructure:"POSTGRES_PORT"`
	APIPort           int    `mapstructure:"API_PORT"`
	PostgresEnableTLS bool   `mapstructure:"POSTGRES_ENABLE_TLS"`
}

// Location returns the configured timezone, which the 08:00 and 15:30 shift boundaries are in.
func (c *Config) Location() (*time.Location, error) {
	loc, err := time.LoadLocation(c.Timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid TIMEZONE %q: %w", c.Timezone, err)
	}
	return loc, nil
}

// Init reads configuration from the environment. A .env file in the working directory or a parent is used if present,
// and real environment variables take precedence over it.
func Init() (*Config, error) {
	v := viper.New()
	v.SetDefault("POSTGRES_PORT", 5432)
	v.SetDefault("POSTGRES_ENABLE_TLS", false)
	v.SetDefault("API_HOST", "127.0.0.1")
	v.SetDefault("API_PORT", 8080)
	v.SetDefault("AUTH_MODE", AuthModeDev)
	v.SetDefault("TIMEZONE", "Europe/Oslo")

	if file, err := path.FindFile(".env"); err == nil {
		v.SetConfigType("env")
		v.SetConfigFile(file)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", file, err)
		}
	}

	// Keys must be bound for Unmarshal to see values that only come from the environment.
	for _, key := range []string{
		"POSTGRES_HOST",
		"POSTGRES_PORT",
		"POSTGRES_USER",
		"POSTGRES_PASSWORD",
		"POSTGRES_DB",
		"POSTGRES_ENABLE_TLS",

		"API_HOST",
		"API_PORT",
		"AUTH_MODE",
		"DEV_USER_EMAIL",
		"OIDC_ISSUER_URL",
		"OIDC_CLIENT_ID",
		"TIMEZONE",
	} {
		if err := v.BindEnv(key); err != nil {
			return nil, fmt.Errorf("failed to bind environment variable %s: %w", key, err)
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	if err := config.validate(); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *Config) validate() error {
	var errs []error
	for name, value := range map[string]string{
		"POSTGRES_HOST":     c.PostgresHostname,
		"POSTGRES_USER":     c.PostgresUsername,
		"POSTGRES_PASSWORD": c.PostgresPassword,
		"POSTGRES_DB":       c.PostgresDatabase,
	} {
		if value == "" {
			errs = append(errs, fmt.Errorf("%s is required", name))
		}
	}

	if c.AuthMode != AuthModeDev && c.AuthMode != AuthModeOIDC {
		errs = append(errs, fmt.Errorf("AUTH_MODE must be %q or %q, got %q", AuthModeDev, AuthModeOIDC, c.AuthMode))
	}
	if c.AuthMode == AuthModeOIDC {
		if c.OIDCIssuerURL == "" {
			errs = append(errs, errors.New("OIDC_ISSUER_URL is required when AUTH_MODE=oidc"))
		}
		if c.OIDCClientID == "" {
			errs = append(errs, errors.New("OIDC_CLIENT_ID is required when AUTH_MODE=oidc"))
		}
	}
	if _, err := c.Location(); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
