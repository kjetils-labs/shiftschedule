package config_test

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"testing"

	"github.com/shiftschedule/internal/config"
	"github.com/shiftschedule/internal/helpers/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {

	err := verifyEnv(t)
	require.Nil(t, err)

	config, err := config.Init()
	require.Nil(t, err)

	assert.NotNil(t, config)

	assert.NotEmpty(t, config.PostgresUsername, "username cannot be empty")
	assert.NotEmpty(t, config.PostgresPassword, "password cannot be empty")
	assert.NotEmpty(t, config.PostgresDatabase, "database cannot be empty")
	assert.NotEmpty(t, config.PostgresHostname, "hostname cannot be empty")
	assert.NotEmpty(t, config.PostgresPort, "port cannot be empty")

}

func verifyEnv(t *testing.T) error {

	envFile := ".env"
	file, err := path.FindFile(envFile)
	require.Nil(t, err)

	_, err = os.ReadFile(file)
	if err != nil {
		t.Logf("failed to read .env. %v", err)
	}

	switch {
	case errors.Is(err, fs.ErrNotExist):
		exampleFile := "./env.example"
		file, err := path.FindFile(exampleFile)
		example, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("failed to stat env.example. %w", err)
		}

		defer func() { err = example.Close() }()
		if err != nil {
			return fmt.Errorf("failed to stat env.example. %w", err)
		}

		newEnv, err := os.Create(envFile)
		if err != nil {
			return fmt.Errorf("failed to create new .env file. %w", err)
		}
		_, err = io.Copy(example, newEnv)
		if err != nil {
			return fmt.Errorf("failed to copy data to .env file. %w", err)
		}

	default:
		return err
	}

	return nil
}
