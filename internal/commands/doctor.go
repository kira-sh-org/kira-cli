package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"kira/internal/config"
	"kira/internal/validation"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check for and fix duplicate work item IDs",
	Long:  `Checks for and fixes duplicate work item IDs by updating the latest one with a new ID.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkWorkDir(); err != nil {
			return err
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		return fixDuplicateIDs(cfg)
	},
}

func fixDuplicateIDs(cfg *config.Config) error {
	result, err := validation.FixDuplicateIDs()
	if err != nil {
		return fmt.Errorf("failed to fix duplicate IDs: %w", err)
	}

	if result.HasErrors() {
		fmt.Println("Issues found and fixed:")
		for _, err := range result.Errors {
			fmt.Printf("  %s\n", err.Error())
		}
	} else {
		fmt.Println("No duplicate IDs found. All work items have unique IDs.")
	}

	return nil
}
