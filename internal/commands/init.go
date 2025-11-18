package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"kira/internal/config"
	"kira/internal/templates"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [folder]",
	Short: "Initialize a kira workspace",
	Long:  `Creates the files and folders used by kira in the specified directory.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := "."
		if len(args) > 0 {
			targetDir = args[0]
		}

		force, _ := cmd.Flags().GetBool("force")
		fillMissing, _ := cmd.Flags().GetBool("fill-missing")
		workPath := filepath.Join(targetDir, ".work")
		if err := ensureDirDecision(workPath, force, fillMissing); err != nil {
			return err
		}

		return initializeWorkspace(targetDir)
	},
}

func init() {
	initCmd.Flags().Bool("force", false, "Overwrite existing .work directory if present")
	initCmd.Flags().Bool("fill-missing", false, "Create any missing files/folders without overwriting existing ones")
}

func initializeWorkspace(targetDir string) error {
	// Create .work directory
	workDir := filepath.Join(targetDir, ".work")
	if err := os.MkdirAll(workDir, 0o700); err != nil {
		return fmt.Errorf("failed to create .work directory: %w", err)
	}

	// Create status folders and .gitkeep files
	cfg := &config.DefaultConfig
	for _, folder := range cfg.StatusFolders {
		folderPath := filepath.Join(workDir, folder)
		if err := os.MkdirAll(folderPath, 0o700); err != nil {
			return fmt.Errorf("failed to create folder %s: %w", folder, err)
		}
		if err := os.WriteFile(filepath.Join(folderPath, ".gitkeep"), []byte(""), 0o600); err != nil {
			return fmt.Errorf("failed to create .gitkeep in %s: %w", folder, err)
		}
	}

	// Create templates directory and default templates and .gitkeep
	if err := templates.CreateDefaultTemplates(workDir); err != nil {
		return fmt.Errorf("failed to create default templates: %w", err)
	}
	if err := os.WriteFile(filepath.Join(workDir, "templates", ".gitkeep"), []byte(""), 0o600); err != nil {
		return fmt.Errorf("failed to create .gitkeep in templates: %w", err)
	}

	// Create or preserve IDEAS.md file (prepend header if missing)
	ideasPath := filepath.Join(workDir, "IDEAS.md")
	header := `# Ideas

This file is for capturing quick ideas and thoughts that don't fit into formal work items yet.

## How to use
- Add ideas with timestamps using ` + "`kira idea \"your idea here\"`" + `
- Or manually add entries below

## List

`
	if _, err := os.Stat(ideasPath); os.IsNotExist(err) {
		if err := os.WriteFile(ideasPath, []byte(header), 0o600); err != nil {
			return fmt.Errorf("failed to create IDEAS.md: %w", err)
		}
	} else {
		content, readErr := os.ReadFile(ideasPath)
		if readErr != nil {
			return fmt.Errorf("failed to read IDEAS.md: %w", readErr)
		}
		if !strings.HasPrefix(string(content), "# Ideas") {
			newContent := header + string(content)
			if err := os.WriteFile(ideasPath, []byte(newContent), 0o600); err != nil {
				return fmt.Errorf("failed to update IDEAS.md: %w", err)
			}
		}
	}

	// Create kira.yml config file under the target directory
	if err := config.SaveConfigToDir(&config.DefaultConfig, targetDir); err != nil {
		return fmt.Errorf("failed to create kira.yml: %w", err)
	}

	fmt.Printf("Initialized kira workspace in %s\n", targetDir)
	return nil
}

func ensureDirDecision(workPath string, force, fillMissing bool) error {
	if _, err := os.Stat(workPath); os.IsNotExist(err) {
		return nil
	}
	if force {
		if err := os.RemoveAll(workPath); err != nil {
			return fmt.Errorf("failed to remove existing .work: %w", err)
		}
		return nil
	}
	if fillMissing {
		return nil
	}

	fmt.Println(".work already exists. Choose an option: [c]ancel, [o]verwrite, [f]ill-missing")
	fmt.Print("Enter choice (c/o/f): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	choice := strings.ToLower(strings.TrimSpace(input))
	switch choice {
	case "o", "overwrite":
		if err := os.RemoveAll(workPath); err != nil {
			return fmt.Errorf("failed to remove existing .work: %w", err)
		}
		return nil
	case "f", "fill-missing":
		return nil
	default:
		return fmt.Errorf("init cancelled")
	}
}
