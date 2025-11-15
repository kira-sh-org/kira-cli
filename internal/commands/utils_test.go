package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindWorkItemFile(t *testing.T) {
	t.Run("finds work item by ID", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir("/")

		// Create .work directory structure
		os.MkdirAll(".work/1_todo", 0755)

		// Create a work item file
		workItemContent := `---
id: 001
title: Test Feature
status: todo
kind: prd
created: 2024-01-01
---

# Test Feature
`
		filePath := ".work/1_todo/001-test-feature.prd.md"
		os.WriteFile(filePath, []byte(workItemContent), 0644)

		// Find the work item
		foundPath, err := findWorkItemFile("001")
		require.NoError(t, err)
		assert.Equal(t, filePath, foundPath)
	})

	t.Run("returns error when work item not found", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir("/")

		// Create .work directory structure
		os.MkdirAll(".work/1_todo", 0755)

		// Try to find non-existent work item
		_, err := findWorkItemFile("999")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "work item with ID 999 not found")
	})
}

func TestUpdateWorkItemStatus(t *testing.T) {
	t.Run("updates status in work item", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir("/")

		// Create a work item file
		workItemContent := `---
id: 001
title: Test Feature
status: todo
kind: prd
created: 2024-01-01
---

# Test Feature
`
		filePath := "test-work-item.md"
		os.WriteFile(filePath, []byte(workItemContent), 0644)

		// Update status
		err := updateWorkItemStatus(filePath, "doing")
		require.NoError(t, err)

		// Check that status was updated
		content, err := os.ReadFile(filePath)
		require.NoError(t, err)

		assert.Contains(t, string(content), "status: doing")
		assert.NotContains(t, string(content), "status: todo")
	})
}

func TestGetWorkItemFiles(t *testing.T) {
	t.Run("finds all work item files in directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir("/")

		// Create directory structure
		os.MkdirAll("test-dir", 0755)

		// Create work item files
		workItem1 := `---
id: 001
title: Test Feature 1
---
# Test Feature 1
`
		workItem2 := `---
id: 002
title: Test Feature 2
---
# Test Feature 2
`
		os.WriteFile("test-dir/001-feature1.md", []byte(workItem1), 0644)
		os.WriteFile("test-dir/002-feature2.md", []byte(workItem2), 0644)
		os.WriteFile("test-dir/not-a-work-item.txt", []byte("not a work item"), 0644)

		// Get work item files
		files, err := getWorkItemFiles("test-dir")
		require.NoError(t, err)

		assert.Len(t, files, 2)
		assert.Contains(t, files, "test-dir/001-feature1.md")
		assert.Contains(t, files, "test-dir/002-feature2.md")
		assert.NotContains(t, files, "test-dir/not-a-work-item.txt")
	})
}

func TestArchiveWorkItems(t *testing.T) {
	t.Run("archives work items to archive directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir("/")

		// Create .work directory structure
		os.MkdirAll(".work", 0755)

		// Create work item files
		workItem1 := `---
id: 001
title: Test Feature 1
---
# Test Feature 1
`
		workItem2 := `---
id: 002
title: Test Feature 2
---
# Test Feature 2
`
		os.WriteFile("work-item1.md", []byte(workItem1), 0644)
		os.WriteFile("work-item2.md", []byte(workItem2), 0644)

		workItems := []string{"work-item1.md", "work-item2.md"}

		// Archive work items
		archivePath, err := archiveWorkItems(workItems, "source-dir")
		require.NoError(t, err)

		// Check that archive directory was created
		assert.DirExists(t, archivePath)

		// Check that work items were copied to archive
		archivedFile1 := filepath.Join(archivePath, "work-item1.md")
		archivedFile2 := filepath.Join(archivePath, "work-item2.md")

		assert.FileExists(t, archivedFile1)
		assert.FileExists(t, archivedFile2)

		// Check that content was preserved
		content1, err := os.ReadFile(archivedFile1)
		require.NoError(t, err)
		assert.Contains(t, string(content1), "Test Feature 1")

		content2, err := os.ReadFile(archivedFile2)
		require.NoError(t, err)
		assert.Contains(t, string(content2), "Test Feature 2")
	})
}
