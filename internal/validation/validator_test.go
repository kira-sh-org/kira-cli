package validation

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"kira/internal/config"
)

func TestValidateWorkItems(t *testing.T) {
	t.Run("validates work items successfully", func(t *testing.T) {
		// Create a temporary workspace
		tmpDir := t.TempDir()
		require.NoError(t, os.Chdir(tmpDir))
		defer func() { _ = os.Chdir("/") }()

		// Create .work directory structure
		require.NoError(t, os.MkdirAll(".work/1_todo", 0o755))

		// Create a valid work item
		workItemContent := `---
id: 001
title: Test Feature
status: todo
kind: prd
created: 2024-01-01
---

# Test Feature

## Context
This is a test feature.
`

		require.NoError(t, os.WriteFile(".work/1_todo/001-test-feature.prd.md", []byte(workItemContent), 0o644))

		cfg := &config.DefaultConfig
		result, err := ValidateWorkItems(cfg)
		require.NoError(t, err)

		assert.False(t, result.HasErrors())
	})

	t.Run("detects missing required fields", func(t *testing.T) {
		// Create a temporary workspace
		tmpDir := t.TempDir()
		require.NoError(t, os.Chdir(tmpDir))
		defer func() { _ = os.Chdir("/") }()

		// Create .work directory structure
		require.NoError(t, os.MkdirAll(".work/1_todo", 0o755))

		// Create an invalid work item (missing title in front matter)
		workItemContent := `---
id: 001
status: todo
kind: prd
created: 2024-01-01
---

# Test Feature
`

		require.NoError(t, os.WriteFile(".work/1_todo/001-test-feature.prd.md", []byte(workItemContent), 0o644))

		cfg := &config.DefaultConfig
		result, err := ValidateWorkItems(cfg)
		require.NoError(t, err)

		// Now that YAML front matter is parsed, missing title should be detected
		assert.True(t, result.HasErrors())
	})
}

func TestGetNextID(t *testing.T) {
	t.Run("generates first ID when no work items exist", func(t *testing.T) {
		// Create a temporary workspace
		tmpDir := t.TempDir()
		require.NoError(t, os.Chdir(tmpDir))
		defer func() { _ = os.Chdir("/") }()

		require.NoError(t, os.MkdirAll(".work", 0o755))

		id, err := GetNextID()
		require.NoError(t, err)
		assert.Equal(t, "001", id)
	})

	t.Run("generates next sequential ID", func(t *testing.T) {
		// Create a temporary workspace
		tmpDir := t.TempDir()
		require.NoError(t, os.Chdir(tmpDir))
		defer func() { _ = os.Chdir("/") }()

		require.NoError(t, os.MkdirAll(".work/1_todo", 0o755))

		// Create a work item with ID 001
		workItemContent := `---
id: 001
title: Test Feature
status: todo
kind: prd
created: 2024-01-01
---

# Test Feature
`

		require.NoError(t, os.WriteFile(".work/1_todo/001-test-feature.prd.md", []byte(workItemContent), 0o644))

		id, err := GetNextID()
		require.NoError(t, err)
		assert.Equal(t, "002", id)
	})
}

func TestFixDuplicateIDs(t *testing.T) {
	t.Run("fixes duplicate IDs", func(t *testing.T) {
		// Create a temporary workspace
		tmpDir := t.TempDir()
		require.NoError(t, os.Chdir(tmpDir))
		defer func() { _ = os.Chdir("/") }()

		require.NoError(t, os.MkdirAll(".work/1_todo", 0o755))

		// Create two work items with the same ID
		workItemContent1 := `---
id: 001
title: First Feature
status: todo
kind: prd
created: 2024-01-01
---

# First Feature
`

		workItemContent2 := `---
id: 001
title: Second Feature
status: todo
kind: prd
created: 2024-01-02
---

# Second Feature
`

		require.NoError(t, os.WriteFile(".work/1_todo/001-first-feature.prd.md", []byte(workItemContent1), 0o644))
		require.NoError(t, os.WriteFile(".work/1_todo/001-second-feature.prd.md", []byte(workItemContent2), 0o644))

		result, err := FixDuplicateIDs()
		require.NoError(t, err)

		// Should not have errors (duplicates should be fixed)
		assert.False(t, result.HasErrors())
	})
}
