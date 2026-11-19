package config_test

import (
	"testing"

	"github.com/shiftschedule/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setRequired(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir()) // keeps a developer's .env out of the tests
	t.Setenv("POSTGRES_HOST", "db.example")
	t.Setenv("POSTGRES_USER", "user")
	t.Setenv("POSTGRES_PASSWORD", "secret")
	t.Setenv("POSTGRES_DB", "shifts")
}

func TestInitFromEnvironmentWithDefaults(t *testing.T) {
	setRequired(t)
	t.Setenv("POSTGRES_PORT", "6543")

	cfg, err := config.Init()
	require.NoError(t, err)

	assert.Equal(t, "db.example", cfg.PostgresHostname)
	assert.Equal(t, 6543, cfg.PostgresPort)
	assert.Equal(t, config.AuthModeDev, cfg.AuthMode)
	assert.Equal(t, 8080, cfg.APIPort)

	loc, err := cfg.Location()
	require.NoError(t, err)
	assert.Equal(t, "Europe/Oslo", loc.String())
}

func TestInitRejectsInvalidValues(t *testing.T) {
	setRequired(t)
	t.Setenv("AUTH_MODE", "magic")
	t.Setenv("TIMEZONE", "Mars/Olympus")

	_, err := config.Init()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "AUTH_MODE")
	assert.Contains(t, err.Error(), "TIMEZONE")
}

func TestInitRequiresIssuerAndClientForOIDC(t *testing.T) {
	setRequired(t)
	t.Setenv("AUTH_MODE", "oidc")

	_, err := config.Init()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OIDC_ISSUER_URL")
	assert.Contains(t, err.Error(), "OIDC_CLIENT_ID")

	t.Setenv("OIDC_ISSUER_URL", "http://127.0.0.1:5556/dex")
	t.Setenv("OIDC_CLIENT_ID", "shiftschedule-web")
	cfg, err := config.Init()
	require.NoError(t, err)
	assert.Equal(t, "http://127.0.0.1:5556/dex", cfg.OIDCIssuerURL)
}
