package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMustLoad(t *testing.T) {
	configContent := `
log_level: INFO
search_concurrency: 10
search_rate: 100
api_server:
  address: "localhost:80"
  timeout: 10s
words_address: "localhost:81"
update_address: "localhost:82"
search_address: "localhost:83"
token_ttl: 120s
`

	tmpFile, err := os.CreateTemp("", "test_config_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	cfg := MustLoad(tmpFile.Name())

	assert.Equal(t, "INFO", cfg.LogLevel)
	assert.Equal(t, 10, cfg.SearchConcurrency)
	assert.Equal(t, 100, cfg.SearchRate)
	assert.Equal(t, "localhost:80", cfg.HTTPConfig.Address)
	assert.Equal(t, 10*time.Second, cfg.HTTPConfig.Timeout)
	assert.Equal(t, "localhost:81", cfg.WordsAddress)
	assert.Equal(t, "localhost:82", cfg.UpdateAddress)
	assert.Equal(t, "localhost:83", cfg.SearchAddress)
	assert.Equal(t, 120*time.Second, cfg.TokenTTL)
}

func TestMustLoad_Defaults(t *testing.T) {
	configContent := "{}"

	tmpFile, err := os.CreateTemp("", "test_config*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	cfg := MustLoad(tmpFile.Name())

	assert.Equal(t, "DEBUG", cfg.LogLevel)
	assert.Equal(t, 1, cfg.SearchConcurrency)
	assert.Equal(t, 1, cfg.SearchRate)
	assert.Equal(t, "localhost:80", cfg.HTTPConfig.Address)
	assert.Equal(t, 5*time.Second, cfg.HTTPConfig.Timeout)
	assert.Equal(t, "words:81", cfg.WordsAddress)
	assert.Equal(t, "update:82", cfg.UpdateAddress)
	assert.Equal(t, "search:83", cfg.SearchAddress)
	assert.Equal(t, 24*time.Hour, cfg.TokenTTL)
}

func TestMustLoad_EnvVars(t *testing.T) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("SEARCH_CONCURRENCY", "10")
	os.Setenv("API_ADDRESS", "localhost:80")
	os.Setenv("TOKEN_TTL", "120s")
	defer func() {
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("SEARCH_CONCURRENCY")
		os.Unsetenv("API_ADDRESS")
		os.Unsetenv("TOKEN_TTL")
	}()

	configContent := "{}"

	tmpFile, err := os.CreateTemp("", "test_config*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	cfg := MustLoad(tmpFile.Name())

	assert.Equal(t, "DEBUG", cfg.LogLevel)
	assert.Equal(t, 10, cfg.SearchConcurrency)
	assert.Equal(t, "localhost:80", cfg.HTTPConfig.Address)
	assert.Equal(t, 120*time.Second, cfg.TokenTTL)
	assert.Equal(t, 1, cfg.SearchRate)
	assert.Equal(t, 5*time.Second, cfg.HTTPConfig.Timeout)
}
