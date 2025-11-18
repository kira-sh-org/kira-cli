package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializeWorkspace(t *testing.T) {
	t.Run("creates workspace structure", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := initializeWorkspace(tmpDir)
		require.NoError(t, err)

		// Check that .work directory was created
		workDir := filepath.Join(tmpDir, ".work")
		assert.DirExists(t, workDir)

		// Check that status folders were created
		statusFolders := []string{"0_backlog", "1_todo", "2_doing", "3_review", "4_done", "z_archive"}
		for _, folder := range statusFolders {
			assert.DirExists(t, filepath.Join(workDir, folder))
		}

		// Check that templates directory was created
		assert.DirExists(t, filepath.Join(workDir, "templates"))

		// Check that IDEAS.md was created
		ideasPath := filepath.Join(workDir, "IDEAS.md")
		assert.FileExists(t, ideasPath)

		// Check that kira.yml was created at root of targetDir
		configPath := filepath.Join(tmpDir, "kira.yml")
		assert.FileExists(t, configPath)
	})

	t.Run("preserves existing files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a pre-existing file
		existingFile := filepath.Join(tmpDir, "existing.txt")
		err := os.WriteFile(existingFile, []byte("existing content"), 0o600)
		require.NoError(t, err)

		err = initializeWorkspace(tmpDir)
		require.NoError(t, err)

		// Check that existing file is still there
		assert.FileExists(t, existingFile)
		content, err := os.ReadFile(existingFile)
		require.NoError(t, err)
		assert.Equal(t, "existing content", string(content))
	})
}

func TestIdeasFileBehavior(t *testing.T) {
	t.Run("prepends header when IDEAS.md exists without header", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Pre-create .work/IDEAS.md with custom content only
		workDir := filepath.Join(tmpDir, ".work")
		require.NoError(t, os.MkdirAll(workDir, 0o700))
		existing := "Custom ideas content\n- [2025-01-01] Something\n"
		require.NoError(t, os.WriteFile(filepath.Join(workDir, "IDEAS.md"), []byte(existing), 0o600))

		// Initialize (should prepend header without wiping existing)
		err := initializeWorkspace(tmpDir)
		require.NoError(t, err)

		data, readErr := os.ReadFile(filepath.Join(workDir, "IDEAS.md"))
		require.NoError(t, readErr)
		content := string(data)
		assert.True(t, strings.HasPrefix(content, "# Ideas"))
		assert.Contains(t, content, "Custom ideas content")
	})

	t.Run("does not duplicate header when re-running", func(t *testing.T) {
		tmpDir := t.TempDir()

		// First run creates header
		require.NoError(t, initializeWorkspace(tmpDir))
		// Second run should not duplicate header
		require.NoError(t, initializeWorkspace(tmpDir))

		data, err := os.ReadFile(filepath.Join(tmpDir, ".work", "IDEAS.md"))
		require.NoError(t, err)
		content := string(data)
		// Count only top-level "# Ideas" lines (ignore "## Ideas")
		lines := strings.Split(content, "\n")
		headerCount := 0
		for _, l := range lines {
			if strings.TrimSpace(l) == "# Ideas" {
				headerCount++
			}
		}
		assert.Equal(t, 1, headerCount)
	})
}
