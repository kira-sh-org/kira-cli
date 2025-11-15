package commands

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddIdea(t *testing.T) {
	t.Run("adds idea to IDEAS.md", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir("/")

		// Create .work directory and IDEAS.md
		os.MkdirAll(".work", 0755)
		ideasContent := `# Ideas

## Ideas

`
		os.WriteFile(".work/IDEAS.md", []byte(ideasContent), 0644)

		// Add an idea
		err := addIdea("Test idea for testing")
		require.NoError(t, err)

		// Check that the idea was added
		content, err := os.ReadFile(".work/IDEAS.md")
		require.NoError(t, err)

		assert.Contains(t, string(content), "Test idea for testing")
		assert.Contains(t, string(content), "## Ideas")
	})

	t.Run("adds timestamp to idea", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir("/")

		// Create .work directory and IDEAS.md
		os.MkdirAll(".work", 0755)
		ideasContent := `# Ideas

## Ideas

`
		os.WriteFile(".work/IDEAS.md", []byte(ideasContent), 0644)

		// Add an idea
		beforeTime := time.Now()
		err := addIdea("Timestamped idea")
		require.NoError(t, err)
		afterTime := time.Now()

		// Check that the idea was added with timestamp
		content, err := os.ReadFile(".work/IDEAS.md")
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, "Timestamped idea")

		// Check that timestamp is in the expected format
		lines := strings.Split(contentStr, "\n")
		var ideaLine string
		for _, line := range lines {
			if strings.Contains(line, "Timestamped idea") {
				ideaLine = line
				break
			}
		}

		assert.NotEmpty(t, ideaLine)
		assert.Contains(t, ideaLine, "- [")
		assert.Contains(t, ideaLine, "]")

		// Parse timestamp and verify it's within expected range
		startIdx := strings.Index(ideaLine, "[") + 1
		endIdx := strings.Index(ideaLine, "]")
		timestampStr := ideaLine[startIdx:endIdx]

		timestamp, err := time.Parse("2006-01-02 15:04:05", timestampStr)
		require.NoError(t, err)

		assert.True(t, timestamp.After(beforeTime.Add(-time.Second)))
		assert.True(t, timestamp.Before(afterTime.Add(time.Second)))
	})
}
