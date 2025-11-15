package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"kira/internal/config"
	"kira/internal/validation"
)

var saveCmd = &cobra.Command{
	Use:   "save [commit-message]",
	Short: "Update work items and commit changes to git",
	Long: `Updates the updated field in work items and commits changes to git.
Validates all non-archived work items before staging.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkWorkDir(); err != nil {
			return err
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		var commitMessage string
		if len(args) > 0 {
			commitMessage = args[0]
		}

		return saveWorkItems(cfg, commitMessage)
	},
}

func saveWorkItems(cfg *config.Config, commitMessage string) error {
	// Validate all work items first
	result, err := validation.ValidateWorkItems(cfg)
	if err != nil {
		return fmt.Errorf("failed to validate work items: %w", err)
	}

	if result.HasErrors() {
		fmt.Println("Validation errors found:")
		for _, err := range result.Errors {
			fmt.Printf("  %s\n", err.Error())
		}
		return fmt.Errorf("validation failed - fix errors before saving")
	}

	// Update timestamps for modified work items
	if err := updateWorkItemTimestamps(); err != nil {
		return fmt.Errorf("failed to update timestamps: %w", err)
	}

	// Check for external changes
	hasExternalChanges, err := checkExternalChanges()
	if err != nil {
		return fmt.Errorf("failed to check for external changes: %w", err)
	}

	if hasExternalChanges {
		fmt.Println("Warning: External changes detected outside .work/ directory.")
		fmt.Println("Skipping commit to avoid mixing work item changes with other changes.")
		return nil
	}

	// Stage only .work/ directory changes
	if err := stageWorkChanges(); err != nil {
		return fmt.Errorf("failed to stage work changes: %w", err)
	}

	// Commit changes
	if commitMessage == "" {
		commitMessage = cfg.Commit.DefaultMessage
	}

	if err := commitChanges(commitMessage); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	fmt.Println("Work items saved and committed successfully.")
	return nil
}

func updateWorkItemTimestamps() error {
	currentTime := time.Now().Format("2006-01-02T15:04:05Z")

	return filepath.Walk(".work", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Skip template files, IDEAS.md, and archived items
		if strings.Contains(path, "template") ||
			strings.HasSuffix(path, "IDEAS.md") ||
			strings.Contains(path, "z_archive") {
			return nil
		}

		// Only process markdown files
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Update the updated timestamp
		return updateFileTimestamp(path, currentTime)
	})
}

func updateFileTimestamp(filePath, timestamp string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	updated := false

	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "updated:") {
			lines[i] = fmt.Sprintf("updated: %s", timestamp)
			updated = true
			break
		}
	}

	// If no updated field found, add it after the created field
	if !updated {
		for i, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "created:") {
				// Insert updated field after created field
				newLines := make([]string, len(lines)+1)
				copy(newLines, lines[:i+1])
				newLines[i+1] = fmt.Sprintf("updated: %s", timestamp)
				copy(newLines[i+2:], lines[i+1:])
				lines = newLines
				break
			}
		}
	}

	return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
}

func checkExternalChanges() (bool, error) {
	// Check git status for changes outside .work/
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		// If git is not available or not a git repo, assume no external changes
		return false, nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "??") {
			// Check if the change is outside .work/
			parts := strings.Fields(line)
			if len(parts) > 1 {
				filePath := parts[1]
				if !strings.HasPrefix(filePath, ".work/") {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func stageWorkChanges() error {
	// Stage all changes in .work/ directory
	cmd := exec.Command("git", "add", ".work/")
	return cmd.Run()
}

func commitChanges(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	return cmd.Run()
}
