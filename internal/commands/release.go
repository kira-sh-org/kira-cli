// Package commands implements the CLI commands for the kira tool.
package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"kira/internal/config"

	"github.com/spf13/cobra"
)

var releaseCmd = &cobra.Command{
	Use:   "release [status|path] [subfolder]",
	Short: "Generate release notes and archive completed work items",
	Long: `Generates release notes and archives completed work items.
Updates work item status to "released" before archival.`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(_ *cobra.Command, args []string) error {
		if err := checkWorkDir(); err != nil {
			return err
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		var targetPath string
		var subfolder string

		if len(args) > 0 {
			targetPath = args[0]
		} else {
			targetPath = "done" // default
		}

		if len(args) > 1 {
			subfolder = args[1]
		}

		return releaseWorkItems(cfg, targetPath, subfolder)
	},
}

func releaseWorkItems(cfg *config.Config, targetPath, subfolder string) error {
	// Determine the source path
	var sourcePath string
	if strings.Contains(targetPath, "/") {
		// Direct path provided
		sourcePath = filepath.Join(".work", targetPath)
	} else {
		// Status name provided
		statusFolder, exists := cfg.StatusFolders[targetPath]
		if !exists {
			return fmt.Errorf("invalid status: %s", targetPath)
		}
		sourcePath = filepath.Join(".work", statusFolder)
	}

	// Add subfolder if provided
	if subfolder != "" {
		sourcePath = filepath.Join(sourcePath, subfolder)
	}

	// Check if source path exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source path does not exist: %s", sourcePath)
	}

	// Get all work item files in the source path
	workItems, err := getWorkItemFiles(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to get work item files: %w", err)
	}

	if len(workItems) == 0 {
		fmt.Println("No work items found to release.")
		return nil
	}

	// Generate release notes
	releaseNotes, err := generateReleaseNotes(workItems)
	if err != nil {
		return fmt.Errorf("failed to generate release notes: %w", err)
	}

	// Update work item statuses to "released"
	for _, workItem := range workItems {
		if err := updateWorkItemStatus(workItem, "released"); err != nil {
			return fmt.Errorf("failed to update work item status: %w", err)
		}
	}

	// Archive work items
	archivePath, err := archiveWorkItems(workItems, sourcePath)
	if err != nil {
		return fmt.Errorf("failed to archive work items: %w", err)
	}

	// Update releases file
	if err := updateReleasesFile(cfg, releaseNotes); err != nil {
		return fmt.Errorf("failed to update releases file: %w", err)
	}

	// Remove original files
	for _, workItem := range workItems {
		if err := os.Remove(workItem); err != nil {
			fmt.Printf("Warning: failed to remove %s: %v\n", workItem, err)
		}
	}

	fmt.Printf("Released %d work items to %s\n", len(workItems), archivePath)
	return nil
}

func generateReleaseNotes(workItems []string) (string, error) {
	var releaseNotes []string

	for _, workItem := range workItems {
		content, err := os.ReadFile(workItem)
		if err != nil {
			return "", err
		}

		// Check if file has release notes section
		if strings.Contains(string(content), "# Release Notes") {
			// Extract release notes section
			lines := strings.Split(string(content), "\n")
			var inReleaseNotes bool
			var releaseNoteLines []string

			for _, line := range lines {
				if strings.Contains(line, "# Release Notes") {
					inReleaseNotes = true
					continue
				}
				if inReleaseNotes {
					if strings.HasPrefix(line, "#") && !strings.Contains(line, "Release Notes") {
						break
					}
					releaseNoteLines = append(releaseNoteLines, line)
				}
			}

			if len(releaseNoteLines) > 0 {
				releaseNotes = append(releaseNotes, strings.Join(releaseNoteLines, "\n"))
			}
		}
	}

	return strings.Join(releaseNotes, "\n\n"), nil
}

func updateReleasesFile(cfg *config.Config, releaseNotes string) error {
	if releaseNotes == "" {
		return nil // No release notes to add
	}

	releasesPath := cfg.Release.ReleasesFile
	var content string

	// Read existing content if file exists
	if _, err := os.Stat(releasesPath); err == nil {
		existing, err := os.ReadFile(releasesPath)
		if err != nil {
			return fmt.Errorf("failed to read releases file: %w", err)
		}
		content = string(existing)
	}

	// Prepend new release notes
	date := time.Now().Format("2006-01-02")
	newContent := fmt.Sprintf("# Release %s\n\n%s\n\n%s", date, releaseNotes, content)

	// Write back to file
	if err := os.WriteFile(releasesPath, []byte(newContent), 0o600); err != nil {
		return fmt.Errorf("failed to write releases file: %w", err)
	}

	return nil
}
