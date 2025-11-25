package test

import "github.com/shiftschedule/internal/config"

func InitTestConfig() *config.Config {

	return &config.Config{
		PostgresHostname:  "127.0.0.1",
		PostgresPort:      5432,
		PostgresUsername:  "testuser",
		PostgresPassword:  "password123",
		PostgresDatabase:  "testDB",
		PostgresEnableTLS: false,
	}
}
