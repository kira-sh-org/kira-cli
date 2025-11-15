package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"kira/internal/config"
)

var abandonCmd = &cobra.Command{
	Use:   "abandon <work-item-id|path> [reason|subfolder]",
	Short: "Archive work items and mark them as abandoned",
	Long: `Archives work items and marks them as abandoned.
Updates work item status to "abandoned" before archival.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkWorkDir(); err != nil {
			return err
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		target := args[0]
		var reasonOrSubfolder string
		if len(args) > 1 {
			reasonOrSubfolder = args[1]
		}

		return abandonWorkItems(cfg, target, reasonOrSubfolder)
	},
}

func abandonWorkItems(cfg *config.Config, target, reasonOrSubfolder string) error {
	var workItems []string
	var sourcePath string

	// Check if target is a work item ID or a path
	if isWorkItemID(target) {
		// Find work item by ID
		workItemPath, err := findWorkItemFile(target)
		if err != nil {
			return err
		}
		workItems = []string{workItemPath}
		sourcePath = filepath.Dir(workItemPath)
	} else {
		// Target is a path
		if strings.Contains(target, "/") {
			sourcePath = filepath.Join(".work", target)
		} else {
			// Status name provided
			statusFolder, exists := cfg.StatusFolders[target]
			if !exists {
				return fmt.Errorf("invalid status: %s", target)
			}
			sourcePath = filepath.Join(".work", statusFolder)
		}

		// Add subfolder if provided
		if reasonOrSubfolder != "" && !strings.Contains(reasonOrSubfolder, " ") {
			// Likely a subfolder (no spaces)
			sourcePath = filepath.Join(sourcePath, reasonOrSubfolder)
		}

		// Get all work item files in the source path
		var err error
		workItems, err = getWorkItemFiles(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to get work item files: %w", err)
		}
	}

	if len(workItems) == 0 {
		fmt.Println("No work items found to abandon.")
		return nil
	}

	// Update work item statuses to "abandoned" and add reason if provided
	for _, workItem := range workItems {
		if err := updateWorkItemStatus(workItem, "abandoned"); err != nil {
			return fmt.Errorf("failed to update work item status: %w", err)
		}

		// Add abandonment reason if provided and it contains spaces (likely a reason)
		if reasonOrSubfolder != "" && strings.Contains(reasonOrSubfolder, " ") {
			if err := addAbandonmentReason(workItem, reasonOrSubfolder); err != nil {
				return fmt.Errorf("failed to add abandonment reason: %w", err)
			}
		}
	}

	// Archive work items
	archivePath, err := archiveWorkItems(workItems, sourcePath)
	if err != nil {
		return fmt.Errorf("failed to archive work items: %w", err)
	}

	// Remove original files
	for _, workItem := range workItems {
		if err := os.Remove(workItem); err != nil {
			fmt.Printf("Warning: failed to remove %s: %v\n", workItem, err)
		}
	}

	fmt.Printf("Abandoned %d work items to %s\n", len(workItems), archivePath)
	return nil
}

func isWorkItemID(target string) bool {
	// Simple check: if it's a 3-digit number, it's likely an ID
	return len(target) == 3 && target[0] >= '0' && target[0] <= '9'
}

func addAbandonmentReason(filePath, reason string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Add abandonment reason to the end of the file
	abandonmentNote := fmt.Sprintf("\n\n## Abandonment\n\n**Reason:** %s\n**Date:** %s\n", reason, time.Now().Format("2006-01-02 15:04:05"))
	newContent := string(content) + abandonmentNote

	return os.WriteFile(filePath, []byte(newContent), 0644)
}
