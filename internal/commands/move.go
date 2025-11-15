package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"kira/internal/config"
)

var moveCmd = &cobra.Command{
	Use:   "move <work-item-id> [target-status]",
	Short: "Move a work item to a different status folder",
	Long:  `Moves the work item to the target status folder. Will display options if target status not provided.`,
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkWorkDir(); err != nil {
			return err
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		workItemID := args[0]
		var targetStatus string
		if len(args) > 1 {
			targetStatus = args[1]
		}

		return moveWorkItem(cfg, workItemID, targetStatus)
	},
}

func moveWorkItem(cfg *config.Config, workItemID, targetStatus string) error {
	// Find the work item file
	workItemPath, err := findWorkItemFile(workItemID)
	if err != nil {
		return err
	}

	// Get target status if not provided
	if targetStatus == "" {
		var err error
		targetStatus, err = selectTargetStatus(cfg)
		if err != nil {
			return err
		}
	}

	// Validate target status
	if _, exists := cfg.StatusFolders[targetStatus]; !exists {
		return fmt.Errorf("invalid target status: %s", targetStatus)
	}

	// Get target folder path
	targetFolder := filepath.Join(".work", cfg.StatusFolders[targetStatus])

	// Move the file
	filename := filepath.Base(workItemPath)
	targetPath := filepath.Join(targetFolder, filename)

	if err := os.Rename(workItemPath, targetPath); err != nil {
		return fmt.Errorf("failed to move work item: %w", err)
	}

	// Update the status in the file
	if err := updateWorkItemStatus(targetPath, targetStatus); err != nil {
		return fmt.Errorf("failed to update work item status: %w", err)
	}

	fmt.Printf("Moved work item %s to %s\n", workItemID, targetStatus)
	return nil
}

func selectTargetStatus(cfg *config.Config) (string, error) {
	fmt.Println("Available statuses:")
	var statuses []string
	for status := range cfg.StatusFolders {
		statuses = append(statuses, status)
	}

	for i, status := range statuses {
		fmt.Printf("%d. %s\n", i+1, status)
	}

	fmt.Print("Select target status (number): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	choice, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || choice < 1 || choice > len(statuses) {
		return "", fmt.Errorf("invalid status selection")
	}

	return statuses[choice-1], nil
}
