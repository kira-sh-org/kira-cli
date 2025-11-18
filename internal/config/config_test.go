package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	t.Run("loads default config when no file exists", func(t *testing.T) {
		// Remove any existing config files
		_ = os.Remove("kira.yml")
		_ = os.Remove(".work/kira.yml")

		config, err := LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "1.0", config.Version)
		assert.NotEmpty(t, config.Templates)
		assert.NotEmpty(t, config.StatusFolders)
	})

	t.Run("loads config from file when exists", func(t *testing.T) {
		// Create a test config file at root-level
		testConfig := `version: "2.0"
templates:
  prd: "custom/prd.md"
status_folders:
  todo: "custom_todo"
`

		require.NoError(t, os.WriteFile("kira.yml", []byte(testConfig), 0o600))
		defer func() { _ = os.Remove("kira.yml") }()

		config, err := LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "2.0", config.Version)
		assert.Equal(t, "custom/prd.md", config.Templates["prd"])
		assert.Equal(t, "custom_todo", config.StatusFolders["todo"])
	})
}

func TestSaveConfig(t *testing.T) {
	t.Run("saves config to file", func(t *testing.T) {
		defer func() { _ = os.Remove("kira.yml") }()

		config := &Config{
			Version: "1.0",
			Templates: map[string]string{
				"prd": "test/prd.md",
			},
		}

		err := SaveConfig(config)
		require.NoError(t, err)

		// Verify file was created at root
		_, err = os.Stat("kira.yml")
		assert.NoError(t, err)
	})
}
