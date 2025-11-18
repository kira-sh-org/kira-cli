package commands

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// findRepoRoot finds the repository root by walking up from the current file
// until it finds go.mod. In GitHub Actions, uses GITHUB_WORKSPACE if available.
func findRepoRoot() (string, error) {
	// Check for GitHub Actions workspace first
	if workspace := os.Getenv("GITHUB_WORKSPACE"); workspace != "" {
		if _, err := os.Stat(filepath.Join(workspace, "go.mod")); err == nil {
			return workspace, nil
		}
	}

	// Fall back to walking up from the test file
	_, thisFile, _, _ := runtime.Caller(1)
	dir := filepath.Dir(thisFile)
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

// buildKiraBinary builds the kira binary for testing.
func buildKiraBinary(t *testing.T, tmpDir string) string {
	repoRoot, err := findRepoRoot()
	require.NoError(t, err, "failed to find repo root")
	outPath := filepath.Join(tmpDir, "kira")
	mainPath := filepath.Join(repoRoot, "cmd", "kira", "main.go")

	// Verify required paths exist
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("main.go does not exist: %s (error: %v)", mainPath, err)
	}

	// Build from repo root directory - Go needs to be in module context
	buildCmd := exec.Command("go", "build", "-o", outPath, "cmd/kira/main.go")
	buildCmd.Dir = repoRoot
	output, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "build failed: %s", string(output))

	return outPath
}

func TestCLIIntegration(t *testing.T) {
	t.Run("full workflow test", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tmpDir))
		defer func() { _ = os.Chdir(originalDir) }()

		// Build the kira binary for testing
		_ = buildKiraBinary(t, tmpDir)
		defer func() { _ = os.Remove("kira") }()

		// Test kira init
		initCmd := exec.Command("./kira", "init")
		output, err := initCmd.CombinedOutput()
		require.NoError(t, err, "init failed: %s", string(output))
		assert.Contains(t, string(output), "Initialized kira workspace")

		// Check that .work directory was created
		assert.DirExists(t, ".work")
		assert.DirExists(t, ".work/1_todo")
		assert.DirExists(t, ".work/2_doing")
		assert.DirExists(t, ".work/3_review")
		assert.DirExists(t, ".work/4_done")
		assert.DirExists(t, ".work/z_archive")
		assert.DirExists(t, ".work/templates")
		assert.FileExists(t, ".work/IDEAS.md")
		assert.FileExists(t, "kira.yml")

		// Ensure .gitkeep files exist in status folders and templates
		gitkeepPaths := []string{
			".work/0_backlog/.gitkeep",
			".work/1_todo/.gitkeep",
			".work/2_doing/.gitkeep",
			".work/3_review/.gitkeep",
			".work/4_done/.gitkeep",
			".work/z_archive/.gitkeep",
			".work/templates/.gitkeep",
		}
		for _, p := range gitkeepPaths {
			assert.FileExists(t, p)
		}

		// Test kira idea
		ideaCmd := exec.Command("./kira", "idea", "Test idea for integration")
		output, err = ideaCmd.CombinedOutput()
		require.NoError(t, err, "idea failed: %s", string(output))
		assert.Contains(t, string(output), "Added idea: Test idea for integration")

		// Check that idea was added to IDEAS.md
		ideasContent, err := os.ReadFile(".work/IDEAS.md")
		require.NoError(t, err)
		assert.Contains(t, string(ideasContent), "Test idea for integration")

		// Test kira lint (should pass with no work items)
		lintCmd := exec.Command("./kira", "lint")
		output, err = lintCmd.CombinedOutput()
		require.NoError(t, err, "lint failed: %s", string(output))
		assert.Contains(t, string(output), "No issues found")

		// Test kira doctor (should pass with no duplicates)
		doctorCmd := exec.Command("./kira", "doctor")
		output, err = doctorCmd.CombinedOutput()
		require.NoError(t, err, "doctor failed: %s", string(output))
		assert.Contains(t, string(output), "No duplicate IDs found")
	})

	t.Run("default status on new without status argument", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.Chdir(tmpDir))
		defer func() { _ = os.Chdir("/") }()

		// Build the kira binary
		_, thisFile, _, _ := runtime.Caller(0)
		repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
		outPath := filepath.Join(tmpDir, "kira")
		buildCmd := exec.Command("go", "build", "-o", outPath, "kira/cmd/kira")
		buildCmd.Dir = repoRoot
		output, err := buildCmd.CombinedOutput()
		require.NoError(t, err, "build failed: %s", string(output))
		defer func() { _ = os.Remove("kira") }()

		// Initialize workspace
		initCmd := exec.Command("./kira", "init")
		output, err = initCmd.CombinedOutput()
		require.NoError(t, err, "init failed: %s", string(output))

		// Create work item without status (should default to backlog)
		// Note: Without --interactive flag, no prompts should appear (default behavior)
		newCmd := exec.Command("./kira", "new", "prd", "Default Status Feature")
		output, err = newCmd.CombinedOutput()
		require.NoError(t, err, "new failed: %s", string(output))

		// Expect file in 0_backlog
		matches, _ := filepath.Glob(".work/0_backlog/*default-status-feature*.prd.md")
		assert.NotEmpty(t, matches, "expected default status file in backlog")
	})

	t.Run("init handles existing .work with --fill-missing and --force", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.Chdir(tmpDir))
		defer func() { _ = os.Chdir("/") }()

		// Build the kira binary
		_, thisFile, _, _ := runtime.Caller(0)
		repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
		outPath := filepath.Join(tmpDir, "kira")
		buildCmd := exec.Command("go", "build", "-o", outPath, "kira/cmd/kira")
		buildCmd.Dir = repoRoot
		output, err := buildCmd.CombinedOutput()
		require.NoError(t, err, "build failed: %s", string(output))
		defer func() { _ = os.Remove("kira") }()

		// Initialize workspace
		initCmd := exec.Command("./kira", "init")
		output, err = initCmd.CombinedOutput()
		require.NoError(t, err, "init failed: %s", string(output))

		// Create a sentinel and remove a folder to simulate missing
		require.NoError(t, os.WriteFile(".work/1_todo/sentinel.txt", []byte("x"), 0o644))
		require.NoError(t, os.RemoveAll(".work/3_review"))

		// Fill missing
		fillCmd := exec.Command("./kira", "init", "--fill-missing")
		output, err = fillCmd.CombinedOutput()
		require.NoError(t, err, "fill-missing failed: %s", string(output))
		assert.FileExists(t, ".work/1_todo/sentinel.txt")
		assert.DirExists(t, ".work/3_review")

		// Force overwrite
		forceCmd := exec.Command("./kira", "init", "--force")
		output, err = forceCmd.CombinedOutput()
		require.NoError(t, err, "force failed: %s", string(output))
		assert.NoFileExists(t, ".work/1_todo/sentinel.txt")
		assert.DirExists(t, ".work/3_review")
		assert.FileExists(t, ".work/3_review/.gitkeep")
	})

	t.Run("work item creation and management", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.Chdir(tmpDir))
		defer func() { _ = os.Chdir("/") }()

		// Build the kira binary for testing using the repo root as working directory
		repoRoot, err := findRepoRoot()
		require.NoError(t, err, "failed to find repo root")
		outPath := filepath.Join(tmpDir, "kira")
		mainPath := filepath.Join(repoRoot, "cmd", "kira", "main.go")
		// Use absolute path and set working directory so Go can find the module
		buildCmd := exec.Command("go", "build", "-o", outPath, mainPath)
		buildCmd.Dir = repoRoot
		buildCmd.Env = append(os.Environ(), "GO111MODULE=on")
		output, err := buildCmd.CombinedOutput()
		require.NoError(t, err, "build failed: %s", string(output))
		defer func() { _ = os.Remove("kira") }()

		// Initialize workspace
		initCmd := exec.Command("./kira", "init")
		output, err = initCmd.CombinedOutput()
		require.NoError(t, err, "init failed: %s", string(output))

		// Create a work item using template input
		// We'll create a simple work item by writing it directly since interactive input is hard to test
		workItemContent := `---
id: 001
title: Test Feature
status: todo
kind: prd
assigned: test@example.com
estimate: 3
created: 2024-01-01
---

# Test Feature

## Context
This is a test feature for integration testing.

## Requirements
- Implement user authentication
- Add login/logout functionality

## Acceptance Criteria
- [ ] User can log in with email/password
- [ ] User can log out
- [ ] Session is maintained across page refreshes

## Implementation Notes
Use JWT tokens for authentication.

## Release Notes
Added user authentication system.
`
		require.NoError(t, os.WriteFile(".work/1_todo/001-test-feature.prd.md", []byte(workItemContent), 0o644))

		// Test kira lint
		lintCmd := exec.Command("./kira", "lint")
		output, err = lintCmd.CombinedOutput()
		require.NoError(t, err, "lint failed: %s", string(output))
		assert.Contains(t, string(output), "No issues found")

		// Test kira move
		moveCmd := exec.Command("./kira", "move", "001", "doing")
		output, err = moveCmd.CombinedOutput()
		require.NoError(t, err, "move failed: %s", string(output))
		assert.Contains(t, string(output), "Moved work item 001 to doing")

		// Check that file was moved
		assert.FileExists(t, ".work/2_doing/001-test-feature.prd.md")
		assert.NoFileExists(t, ".work/1_todo/001-test-feature.prd.md")

		// Test kira save (this will fail if git is not initialized, which is expected)
		saveCmd := exec.Command("./kira", "save", "Test commit")
		_, err = saveCmd.CombinedOutput()

		// We expect this to fail because git is not initialized
		assert.Error(t, err)
	})

	t.Run("template system test", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.Chdir(tmpDir))
		defer func() { _ = os.Chdir("/") }()

		// Build the kira binary for testing using the repo root as working directory
		repoRoot, err := findRepoRoot()
		require.NoError(t, err, "failed to find repo root")
		outPath := filepath.Join(tmpDir, "kira")
		mainPath := filepath.Join(repoRoot, "cmd", "kira", "main.go")
		// Use absolute path and set working directory so Go can find the module
		buildCmd := exec.Command("go", "build", "-o", outPath, mainPath)
		buildCmd.Dir = repoRoot
		buildCmd.Env = append(os.Environ(), "GO111MODULE=on")
		output, err := buildCmd.CombinedOutput()
		require.NoError(t, err, "build failed: %s", string(output))
		defer func() { _ = os.Remove("kira") }()

		// Initialize workspace
		initCmd := exec.Command("./kira", "init")
		output, err = initCmd.CombinedOutput()
		require.NoError(t, err, "init failed: %s", string(output))

		// Check that templates were created
		templateFiles := []string{
			".work/templates/template.prd.md",
			".work/templates/template.issue.md",
			".work/templates/template.spike.md",
			".work/templates/template.task.md",
		}

		for _, templateFile := range templateFiles {
			assert.FileExists(t, templateFile)

			// Check that template contains input placeholders
			content, err := os.ReadFile(templateFile)
			require.NoError(t, err)
			assert.Contains(t, string(content), "<!--input-")
		}

		// Test help-inputs command
		helpCmd := exec.Command("./kira", "new", "prd", "--help-inputs")
		output, err = helpCmd.CombinedOutput()
		require.NoError(t, err, "help-inputs failed: %s", string(output))
		assert.Contains(t, string(output), "Available inputs for template 'prd'")
	})

	t.Run("release command generates notes and archives done items", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.Chdir(tmpDir))
		defer func() { _ = os.Chdir("/") }()

		// Build the kira binary
		_, thisFile, _, _ := runtime.Caller(0)
		repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
		outPath := filepath.Join(tmpDir, "kira")
		buildCmd := exec.Command("go", "build", "-o", outPath, "kira/cmd/kira")
		buildCmd.Dir = repoRoot
		output, err := buildCmd.CombinedOutput()
		require.NoError(t, err, "build failed: %s", string(output))
		defer func() { _ = os.Remove("kira") }()

		// Initialize workspace
		initCmd := exec.Command("./kira", "init")
		output, err = initCmd.CombinedOutput()
		require.NoError(t, err, "init failed: %s", string(output))

		// Create a done item with Release Notes section
		require.NoError(t, os.MkdirAll(".work/4_done", 0o755))
		doneItem := `---
id: 001
title: Done Feature
status: done
kind: prd
created: 2024-01-01
---

# Done Feature

## Context
Something

## Release Notes
This is a release note entry.
`
		require.NoError(t, os.WriteFile(".work/4_done/001-done-feature.prd.md", []byte(doneItem), 0o644))

		// Run release (default from done)
		releaseCmd := exec.Command("./kira", "release")
		output, err = releaseCmd.CombinedOutput()
		require.NoError(t, err, "release failed: %s", string(output))
		assert.Contains(t, string(output), "Released 1 work items")

		// Check archived file exists and original removed
		archivedMatches, _ := filepath.Glob(".work/z_archive/*/4_done/001-done-feature.prd.md")
		assert.NotEmpty(t, archivedMatches, "archived file not found")
		assert.NoFileExists(t, ".work/4_done/001-done-feature.prd.md")

		// Check RELEASES.md contains note
		releasesContent, err := os.ReadFile("RELEASES.md")
		require.NoError(t, err)
		assert.Contains(t, string(releasesContent), "This is a release note entry.")
	})

	t.Run("abandon command handles id, reason, status path and subfolder", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.Chdir(tmpDir))
		defer func() { _ = os.Chdir("/") }()

		// Build the kira binary
		_, thisFile, _, _ := runtime.Caller(0)
		repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
		outPath := filepath.Join(tmpDir, "kira")
		buildCmd := exec.Command("go", "build", "-o", outPath, "kira/cmd/kira")
		buildCmd.Dir = repoRoot
		output, err := buildCmd.CombinedOutput()
		require.NoError(t, err, "build failed: %s", string(output))
		defer func() { _ = os.Remove("kira") }()

		// Initialize workspace
		initCmd := exec.Command("./kira", "init")
		output, err = initCmd.CombinedOutput()
		require.NoError(t, err, "init failed: %s", string(output))

		// Create two items in todo and a subfolder
		require.NoError(t, os.MkdirAll(".work/1_todo/sub", 0o755))
		item1 := `---
id: 001
title: Todo One
status: todo
kind: prd
created: 2024-01-01
---
`
		item2 := `---
id: 002
title: Todo Two
status: todo
kind: prd
created: 2024-01-01
---
`
		require.NoError(t, os.WriteFile(".work/1_todo/001-todo-one.prd.md", []byte(item1), 0o644))
		require.NoError(t, os.WriteFile(".work/1_todo/sub/002-todo-two.prd.md", []byte(item2), 0o644))

		// Abandon by id with reason
		abandonID := exec.Command("./kira", "abandon", "001", "No longer needed")
		output, err = abandonID.CombinedOutput()
		require.NoError(t, err, "abandon by id failed: %s", string(output))
		// Verify archived
		archived1, _ := filepath.Glob(".work/z_archive/*/1_todo/001-todo-one.prd.md")
		assert.NotEmpty(t, archived1)
		assert.NoFileExists(t, ".work/1_todo/001-todo-one.prd.md")
		// Verify abandonment note present
		content, err := os.ReadFile(archived1[0])
		require.NoError(t, err)
		assert.Contains(t, string(content), "## Abandonment")
		assert.Contains(t, string(content), "No longer needed")

		// Abandon by status path subfolder (todo sub)
		abandonFolder := exec.Command("./kira", "abandon", "todo", "sub")
		output, err = abandonFolder.CombinedOutput()
		require.NoError(t, err, "abandon folder failed: %s", string(output))
		archived2, _ := filepath.Glob(".work/z_archive/*/sub/002-todo-two.prd.md")
		assert.NotEmpty(t, archived2)
		assert.NoFileExists(t, ".work/1_todo/sub/002-todo-two.prd.md")
	})

	t.Run("save command commits and updates timestamps in git repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.Chdir(tmpDir))
		defer func() { _ = os.Chdir("/") }()

		// Build binary
		_, thisFile, _, _ := runtime.Caller(0)
		repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
		outPath := filepath.Join(tmpDir, "kira")
		buildCmd := exec.Command("go", "build", "-o", outPath, "kira/cmd/kira")
		buildCmd.Dir = repoRoot
		output, err := buildCmd.CombinedOutput()
		require.NoError(t, err, "build failed: %s", string(output))
		defer func() { _ = os.Remove("kira") }()

		// git init
		require.NoError(t, exec.Command("git", "init").Run())
		require.NoError(t, exec.Command("git", "config", "user.email", "test@example.com").Run())
		require.NoError(t, exec.Command("git", "config", "user.name", "Test User").Run())

		// Initialize workspace and add initial commit
		initCmd := exec.Command("./kira", "init")
		output, err = initCmd.CombinedOutput()
		require.NoError(t, err, "init failed: %s", string(output))
		require.NoError(t, exec.Command("git", "add", ".").Run())
		require.NoError(t, exec.Command("git", "commit", "-m", "init").Run())

		// Create a work item
		item := `---
id: 001
title: Save Test
status: todo
kind: prd
created: 2024-01-01
---

# Save Test
`
		require.NoError(t, os.WriteFile(".work/1_todo/001-save-test.prd.md", []byte(item), 0o644))

		// Run save with custom message
		saveCmd := exec.Command("./kira", "save", "Custom commit message")
		output, err = saveCmd.CombinedOutput()
		require.NoError(t, err, "save failed: %s", string(output))
		assert.Contains(t, string(output), "Work items saved and committed successfully.")

		// Ensure updated field added
		content, err := os.ReadFile(".work/1_todo/001-save-test.prd.md")
		require.NoError(t, err)
		assert.Contains(t, string(content), "updated:")

		// Verify last commit message
		logOut, err := exec.Command("git", "log", "-1", "--pretty=%B").Output()
		require.NoError(t, err)
		assert.Contains(t, string(logOut), "Custom commit message")

		// Verify only .work files in commit
		showOut, err := exec.Command("git", "show", "--name-only", "--pretty=", "HEAD").Output()
		require.NoError(t, err)
		for _, line := range strings.Split(strings.TrimSpace(string(showOut)), "\n") {
			if line == "" {
				continue
			}
			assert.True(t, strings.HasPrefix(line, ".work/"), "commit touched non-.work file: %s", line)
		}
	})

	t.Run("save fails on validation errors", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.Chdir(tmpDir))
		defer func() { _ = os.Chdir("/") }()

		// Build binary
		_, thisFile, _, _ := runtime.Caller(0)
		repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
		outPath := filepath.Join(tmpDir, "kira")
		buildCmd := exec.Command("go", "build", "-o", outPath, "kira/cmd/kira")
		buildCmd.Dir = repoRoot
		output, err := buildCmd.CombinedOutput()
		require.NoError(t, err, "build failed: %s", string(output))
		defer func() { _ = os.Remove("kira") }()

		// Init workspace (no git init to force no commit path)
		initCmd := exec.Command("./kira", "init")
		_, err = initCmd.CombinedOutput()
		require.NoError(t, err)

		// Create invalid item
		invalid := `---
id: 001
title: Bad
status: invalid-status
kind: prd
created: 2024-01-01
---
`
		require.NoError(t, os.WriteFile(".work/1_todo/001-bad.prd.md", []byte(invalid), 0o644))

		// Save should fail
		saveCmd := exec.Command("./kira", "save", "attempt")
		output, err = saveCmd.CombinedOutput()
		assert.Error(t, err)
		assert.Contains(t, string(output), "validation failed")
	})
}
