// Package commands implements the CLI commands for the kira tool.
package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var ideaCmd = &cobra.Command{
	Use:   "idea <description>",
	Short: "Add an idea to IDEAS.md",
	Long:  `Adds an idea with a timestamp to the IDEAS.md file.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		if err := checkWorkDir(); err != nil {
			return err
		}

		description := args[0]
		return addIdea(description)
	},
}

func addIdea(description string) error {
	ideasPath := filepath.Join(".work", "IDEAS.md")

	// Read existing content
	content, err := os.ReadFile(ideasPath)
	if err != nil {
		return fmt.Errorf("failed to read IDEAS.md: %w", err)
	}

	// Append new idea with timestamp (UTC for deterministic tests)
	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05")
	newIdea := fmt.Sprintf("- [%s] %s\n", timestamp, description)

	// Append to content
	newContent := string(content) + newIdea

	// Write back to file
	if err := os.WriteFile(ideasPath, []byte(newContent), 0o600); err != nil {
		return fmt.Errorf("failed to write IDEAS.md: %w", err)
	}

	fmt.Printf("Added idea: %s\n", description)
	return nil
}
