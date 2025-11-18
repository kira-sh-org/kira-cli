package commands

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"kira/internal/config"
)

func TestLintWorkItems(t *testing.T) {
	t.Run("reports no issues for valid work items", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.Chdir(tmpDir))
		defer func() { _ = os.Chdir("/") }()

		// Create .work directory structure
		require.NoError(t, os.MkdirAll(".work/1_todo", 0o700))

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
		require.NoError(t, os.WriteFile(".work/1_todo/001-test-feature.prd.md", []byte(workItemContent), 0o600))

		cfg := &config.DefaultConfig
		err := lintWorkItems(cfg)
		require.NoError(t, err)
	})

	t.Run("reports validation errors", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.Chdir(tmpDir))
		defer func() { _ = os.Chdir("/") }()

		// Create .work directory structure
		require.NoError(t, os.MkdirAll(".work/1_todo", 0o700))

		// Create an invalid work item (invalid status)
		workItemContent := `---
id: 001
title: Test Feature
status: invalid-status
kind: prd
created: 2024-01-01
---

# Test Feature
`
		require.NoError(t, os.WriteFile(".work/1_todo/001-test-feature.prd.md", []byte(workItemContent), 0o600))

		cfg := &config.DefaultConfig
		err := lintWorkItems(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})
}
