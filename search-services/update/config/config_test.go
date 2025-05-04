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
update_address: localhost:83
xkcd:
  url: xkcd.com
  concurrency: 10
  timeout: 10s
  check_period: 1h
db_address: localhost:82
words_address: localhost:81
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
	assert.Equal(t, "localhost:83", cfg.Address)

	assert.Equal(t, "xkcd.com", cfg.XKCD.URL)
	assert.Equal(t, 10, cfg.XKCD.Concurrency)
	assert.Equal(t, 10*time.Second, cfg.XKCD.Timeout)
	assert.Equal(t, 1*time.Hour, cfg.XKCD.CheckPeriod)

	assert.Equal(t, "localhost:82", cfg.DBAddress)
	assert.Equal(t, "localhost:81", cfg.WordsAddress)
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
	assert.Equal(t, "localhost:83", cfg.Address)
	assert.Equal(t, "localhost:82", cfg.DBAddress)
	assert.Equal(t, "localhost:81", cfg.WordsAddress)
}

func TestMustLoad_EnvVars(t *testing.T) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("UPDATE_ADDRESS", "localhost:83")
	os.Setenv("DB_ADDRESS", "localhost:82")
	os.Setenv("WORDS_ADDRESS", "localhost:81")
	defer func() {
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("SEARCH_ADDRESS")
		os.Unsetenv("DB_ADDRESS")
		os.Unsetenv("WORDS_ADDRESS")
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
	assert.Equal(t, "localhost:83", cfg.Address)
	assert.Equal(t, "localhost:82", cfg.DBAddress)
	assert.Equal(t, "localhost:81", cfg.WordsAddress)
}
