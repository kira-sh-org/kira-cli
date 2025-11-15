package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"kira/internal/config"
	"kira/internal/validation"
)

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Check for issues in work items",
	Long:  `Scans folders and files to check for issues and reports any found.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkWorkDir(); err != nil {
			return err
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		return lintWorkItems(cfg)
	},
}

func lintWorkItems(cfg *config.Config) error {
	result, err := validation.ValidateWorkItems(cfg)
	if err != nil {
		return fmt.Errorf("failed to validate work items: %w", err)
	}

	if result.HasErrors() {
		fmt.Println("Validation errors found:")
		for _, err := range result.Errors {
			fmt.Printf("  %s\n", err.Error())
		}
		return fmt.Errorf("validation failed")
	}

	fmt.Println("No issues found. All work items are valid.")
	return nil
}
