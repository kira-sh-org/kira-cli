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
		os.Chdir(tmpDir)
		defer os.Chdir("/")

		// Create .work directory structure
		os.MkdirAll(".work/1_todo", 0755)

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
		os.WriteFile(".work/1_todo/001-test-feature.prd.md", []byte(workItemContent), 0644)

		cfg := &config.DefaultConfig
		err := lintWorkItems(cfg)
		require.NoError(t, err)
	})

	t.Run("reports validation errors", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir("/")

		// Create .work directory structure
		os.MkdirAll(".work/1_todo", 0755)

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
		os.WriteFile(".work/1_todo/001-test-feature.prd.md", []byte(workItemContent), 0644)

		cfg := &config.DefaultConfig
		err := lintWorkItems(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})
}
