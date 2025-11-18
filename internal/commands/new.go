package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"kira/internal/config"
	"kira/internal/templates"
	"kira/internal/validation"

	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new [template] [status] [title] [description]",
	Short: "Create a new work item",
	Long: `Creates a new work item from a template in the specified status folder.
All arguments are optional - will prompt for selection if not provided.`,
	Args: cobra.MaximumNArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkWorkDir(); err != nil {
			return err
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		interactive, _ := cmd.Flags().GetBool("interactive")
		inputValues, _ := cmd.Flags().GetStringToString("input")
		helpInputs, _ := cmd.Flags().GetBool("help-inputs")

		return createWorkItem(cfg, args, interactive, inputValues, helpInputs)
	},
}

func init() {
	newCmd.Flags().BoolP("interactive", "I", false, "Enable interactive input prompts for missing template fields")
	newCmd.Flags().StringToStringP("input", "i", nil, "Provide input values directly (e.g., --input due=2025-10-01)")
	newCmd.Flags().Bool("help-inputs", false, "List available input variables for a template")
}

func createWorkItem(cfg *config.Config, args []string, interactive bool, inputValues map[string]string, helpInputs bool) error {
	parsedArgs, err := parseWorkItemArgs(cfg, args)
	if err != nil {
		return err
	}

	template, err := resolveTemplate(cfg, parsedArgs.template, helpInputs)
	if err != nil {
		return err
	}

	if helpInputs {
		return showTemplateInputs(cfg, template)
	}

	title, err := resolveTitle(parsedArgs.title, interactive)
	if err != nil {
		return err
	}

	status, err := resolveStatus(cfg, parsedArgs.status)
	if err != nil {
		return err
	}

	nextID, err := validation.GetNextID()
	if err != nil {
		return fmt.Errorf("failed to get next ID: %w", err)
	}

	inputs, err := collectInputs(cfg, template, nextID, title, status, parsedArgs.description, inputValues, interactive)
	if err != nil {
		return err
	}

	return writeWorkItemFile(cfg, template, nextID, title, status, inputs)
}

type workItemArgs struct {
	template    string
	title       string
	status      string
	description string
}

func parseWorkItemArgs(cfg *config.Config, args []string) (workItemArgs, error) {
	var result workItemArgs

	if len(args) > 0 {
		result.template = args[0]
	}

	statusSet := buildStatusSet(cfg)
	validStatuses := buildValidStatuses(cfg)

	if len(args) > 1 {
		if _, ok := statusSet[args[1]]; ok {
			result.status = args[1]
		} else {
			result.title = args[1]
		}
	}

	if len(args) > 2 {
		if err := parseThirdArg(&result, args[2], statusSet, validStatuses); err != nil {
			return result, err
		}
	}

	if len(args) > 3 {
		result.description = args[3]
	}

	return result, nil
}

func buildStatusSet(cfg *config.Config) map[string]struct{} {
	statusSet := make(map[string]struct{}, len(cfg.StatusFolders))
	for s := range cfg.StatusFolders {
		statusSet[s] = struct{}{}
	}
	return statusSet
}

func buildValidStatuses(cfg *config.Config) []string {
	validStatuses := make([]string, 0, len(cfg.StatusFolders))
	for s := range cfg.StatusFolders {
		validStatuses = append(validStatuses, s)
	}
	return validStatuses
}

func parseThirdArg(result *workItemArgs, arg string, statusSet map[string]struct{}, validStatuses []string) error {
	if result.status == "" {
		if _, ok := statusSet[arg]; ok {
			result.status = arg
		} else if result.title == "" {
			result.title = arg
		} else {
			return fmt.Errorf("invalid status: neither '%s' nor '%s' is a valid status (valid: %s)", result.title, arg, strings.Join(validStatuses, ", "))
		}
	} else if result.title == "" {
		result.title = arg
	}
	return nil
}

func resolveTemplate(cfg *config.Config, template string, helpInputs bool) (string, error) {
	if template == "" {
		if helpInputs {
			return "", fmt.Errorf("template must be specified when using --help-inputs")
		}
		return selectTemplate(cfg)
	}
	return template, nil
}

func resolveTitle(title string, interactive bool) (string, error) {
	if title == "" {
		if interactive {
			return promptString("Enter work item title: ")
		}
		return "", fmt.Errorf("title is required (provide as argument or use --interactive flag)")
	}
	return title, nil
}

func resolveStatus(cfg *config.Config, status string) (string, error) {
	if status == "" {
		status = cfg.DefaultStatus
	}
	if _, ok := cfg.StatusFolders[status]; !ok {
		validStatuses := buildValidStatuses(cfg)
		return "", fmt.Errorf("invalid status '%s' (valid: %s)", status, strings.Join(validStatuses, ", "))
	}
	return status, nil
}

func collectInputs(cfg *config.Config, template, nextID, title, status, description string, inputValues map[string]string, interactive bool) (map[string]string, error) {
	inputs := make(map[string]string)
	inputs["id"] = nextID
	inputs["title"] = title
	inputs["status"] = status
	inputs["created"] = time.Now().Format("2006-01-02")

	if description != "" {
		if _, exists := inputValues["description"]; !exists {
			inputValues["description"] = description
		}
	}

	for k, v := range inputValues {
		inputs[k] = v
	}

	if interactive {
		if err := collectInteractiveInputs(cfg, template, inputs); err != nil {
			return nil, err
		}
	}

	return inputs, nil
}

func collectInteractiveInputs(cfg *config.Config, template string, inputs map[string]string) error {
	templatePath := filepath.Join(".work", cfg.Templates[template])
	templateInputs, err := templates.GetTemplateInputs(templatePath)
	if err != nil {
		return fmt.Errorf("failed to get template inputs: %w", err)
	}

	for _, input := range templateInputs {
		if _, exists := inputs[input.Name]; !exists {
			value, err := promptForInput(input)
			if err != nil {
				return err
			}
			inputs[input.Name] = value
		}
	}
	return nil
}

func writeWorkItemFile(cfg *config.Config, template, nextID, title, status string, inputs map[string]string) error {
	templatePath := filepath.Join(".work", cfg.Templates[template])
	content, err := templates.ProcessTemplate(templatePath, inputs)
	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	filename := fmt.Sprintf("%s-%s.%s.md", nextID, kebabCase(title), template)
	statusFolder, exists := cfg.StatusFolders[status]
	if !exists || statusFolder == "" {
		return fmt.Errorf("invalid status folder for status '%s'", status)
	}

	statusFolderPath := filepath.Join(".work", statusFolder)
	if err := os.MkdirAll(statusFolderPath, 0o700); err != nil {
		return fmt.Errorf("failed to create status folder: %w", err)
	}

	filePath := filepath.Join(statusFolderPath, filename)
	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("failed to write work item file: %w", err)
	}

	fmt.Printf("Created work item %s in %s\n", nextID, statusFolder)
	return nil
}

func selectTemplate(cfg *config.Config) (string, error) {
	fmt.Println("Available templates:")
	var templates []string
	for template := range cfg.Templates {
		templates = append(templates, template)
	}

	for i, template := range templates {
		fmt.Printf("%d. %s\n", i+1, template)
	}

	fmt.Print("Select template (number): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	choice, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || choice < 1 || choice > len(templates) {
		return "", fmt.Errorf("invalid template selection")
	}

	return templates[choice-1], nil
}

func showTemplateInputs(cfg *config.Config, template string) error {
	templatePath := filepath.Join(".work", cfg.Templates[template])
	inputs, err := templates.GetTemplateInputs(templatePath)
	if err != nil {
		return fmt.Errorf("failed to get template inputs: %w", err)
	}

	fmt.Printf("Available inputs for template '%s':\n", template)
	for _, input := range inputs {
		fmt.Printf("- %s (%s): %s\n", input.Name, input.Type, input.Description)
		if len(input.Options) > 0 {
			fmt.Printf("  Options: %s\n", strings.Join(input.Options, ", "))
		}
	}

	return nil
}

func promptForInput(input templates.Input) (string, error) {
	prompt := fmt.Sprintf("Enter %s (%s): ", input.Name, input.Description)

	switch input.Type {
	case templates.InputString:
		if len(input.Options) > 0 {
			return promptStringOptions(prompt, input.Options)
		}
		return promptString(prompt)
	case templates.InputNumber:
		return promptNumber(prompt)
	case templates.InputDateTime:
		return promptDateTime(prompt, input.DateFormat)
	default:
		return promptString(prompt)
	}
}

func promptString(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func promptStringOptions(prompt string, options []string) (string, error) {
	fmt.Println(prompt)
	for i, option := range options {
		fmt.Printf("%d. %s\n", i+1, option)
	}
	fmt.Print("Select option (number): ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	choice, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || choice < 1 || choice > len(options) {
		return "", fmt.Errorf("invalid option selection")
	}

	return options[choice-1], nil
}

func promptNumber(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	// Validate it's a number
	_, err = strconv.Atoi(strings.TrimSpace(input))
	if err != nil {
		return "", fmt.Errorf("invalid number: %v", err)
	}

	return strings.TrimSpace(input), nil
}

func promptDateTime(prompt, format string) (string, error) {
	fmt.Printf("%s (format: %s): ", prompt, format)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	// Validate date format
	_, err = time.Parse(format, strings.TrimSpace(input))
	if err != nil {
		return "", fmt.Errorf("invalid date format: %v", err)
	}

	return strings.TrimSpace(input), nil
}

func kebabCase(s string) string {
	// Simple kebab case conversion
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	return s
}
