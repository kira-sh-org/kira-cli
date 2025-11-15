package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v3"

	"kira/internal/config"
)

type ValidationError struct {
	File    string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.File, e.Message)
}

type ValidationResult struct {
	Errors []ValidationError
}

func (r *ValidationResult) AddError(file, message string) {
	r.Errors = append(r.Errors, ValidationError{File: file, Message: message})
}

func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

func (r *ValidationResult) Error() string {
	if !r.HasErrors() {
		return ""
	}

	var messages []string
	for _, err := range r.Errors {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "\n")
}

type WorkItem struct {
	ID      string                 `yaml:"id"`
	Title   string                 `yaml:"title"`
	Status  string                 `yaml:"status"`
	Kind    string                 `yaml:"kind"`
	Created string                 `yaml:"created"`
	Fields  map[string]interface{} `yaml:",inline"`
}

func ValidateWorkItems(cfg *config.Config) (*ValidationResult, error) {
	result := &ValidationResult{}

	// Get all work item files
	files, err := getWorkItemFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get work item files: %w", err)
	}

	// Track IDs for duplicate checking
	idMap := make(map[string][]string)

	for _, file := range files {
		workItem, err := parseWorkItemFile(file)
		if err != nil {
			result.AddError(file, fmt.Sprintf("failed to parse file: %v", err))
			continue
		}

		// Validate required fields
		if err := validateRequiredFields(workItem, cfg); err != nil {
			result.AddError(file, err.Error())
		}

		// Validate ID format
		if err := validateIDFormat(workItem.ID, cfg); err != nil {
			result.AddError(file, err.Error())
		}

		// Validate status values
		if err := validateStatus(workItem.Status, cfg); err != nil {
			result.AddError(file, err.Error())
		}

		// Validate date formats
		if err := validateDateFormats(workItem); err != nil {
			result.AddError(file, err.Error())
		}

		// Track ID for duplicate checking
		idMap[workItem.ID] = append(idMap[workItem.ID], file)
	}

	// Check for duplicate IDs
	for id, files := range idMap {
		if len(files) > 1 {
			result.AddError(files[0], fmt.Sprintf("duplicate ID found: %s in files %s", id, strings.Join(files, ", ")))
		}
	}

	// Validate workflow rules
	if err := validateWorkflowRules(cfg); err != nil {
		result.AddError("workflow", err.Error())
	}

	return result, nil
}

func getWorkItemFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(".work", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Skip template files and IDEAS.md
		if strings.Contains(path, "template") || strings.HasSuffix(path, "IDEAS.md") {
			return nil
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

func parseWorkItemFile(filePath string) (*WorkItem, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Extract YAML front matter between the first pair of --- lines
	lines := strings.Split(string(content), "\n")
	var yamlLines []string
	inYAML := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if i == 0 && trimmed == "---" {
			inYAML = true
			continue
		}
		if inYAML {
			if trimmed == "---" {
				break
			}
			yamlLines = append(yamlLines, line)
		}
	}

	wi := &WorkItem{Fields: make(map[string]interface{})}
	if len(yamlLines) > 0 {
		if err := yaml.Unmarshal([]byte(strings.Join(yamlLines, "\n")), wi); err != nil {
			return nil, fmt.Errorf("failed to parse front matter: %w", err)
		}
	}

	return wi, nil
}

func validateRequiredFields(workItem *WorkItem, cfg *config.Config) error {
	for _, field := range cfg.Validation.RequiredFields {
		switch field {
		case "id":
			if workItem.ID == "" {
				return fmt.Errorf("missing required field: id")
			}
		case "title":
			if workItem.Title == "" {
				return fmt.Errorf("missing required field: title")
			}
		case "status":
			if workItem.Status == "" {
				return fmt.Errorf("missing required field: status")
			}
		case "kind":
			if workItem.Kind == "" {
				return fmt.Errorf("missing required field: kind")
			}
		case "created":
			if workItem.Created == "" {
				return fmt.Errorf("missing required field: created")
			}
		}
	}
	return nil
}

func validateIDFormat(id string, cfg *config.Config) error {
	matched, err := regexp.MatchString(cfg.Validation.IDFormat, id)
	if err != nil {
		return fmt.Errorf("invalid ID format regex: %w", err)
	}
	if !matched {
		return fmt.Errorf("invalid ID format: %s (expected format: %s)", id, cfg.Validation.IDFormat)
	}
	return nil
}

func validateStatus(status string, cfg *config.Config) error {
	for _, validStatus := range cfg.Validation.StatusValues {
		if status == validStatus {
			return nil
		}
	}
	return fmt.Errorf("invalid status '%s'. Valid values: %s", status, strings.Join(cfg.Validation.StatusValues, ", "))
}

func validateDateFormats(workItem *WorkItem) error {
	// Validate created date
	if workItem.Created != "" {
		if _, err := time.Parse("2006-01-02", workItem.Created); err != nil {
			return fmt.Errorf("invalid created date format: %s", workItem.Created)
		}
	}

	// Validate other date fields if present
	for key, value := range workItem.Fields {
		if strings.Contains(key, "date") || strings.Contains(key, "due") {
			if str, ok := value.(string); ok && str != "" {
				if _, err := time.Parse("2006-01-02", str); err != nil {
					return fmt.Errorf("invalid %s date format: %s", key, str)
				}
			}
		}
	}

	return nil
}

func validateWorkflowRules(cfg *config.Config) error {
	// Check that only one item is in doing folder
	doingPath := filepath.Join(".work", cfg.StatusFolders["doing"])
	if _, err := os.Stat(doingPath); err == nil {
		files, err := os.ReadDir(doingPath)
		if err != nil {
			return fmt.Errorf("failed to read doing folder: %w", err)
		}

		var workItems []string
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
				workItems = append(workItems, file.Name())
			}
		}

		if len(workItems) > 1 {
			return fmt.Errorf("multiple items in doing folder. Only one item allowed at a time. Found: %s", strings.Join(workItems, ", "))
		}
	}

	return nil
}

func GetNextID() (string, error) {
	files, err := getWorkItemFiles()
	if err != nil {
		return "", fmt.Errorf("failed to get work item files: %w", err)
	}

	var maxID int
	for _, file := range files {
		workItem, err := parseWorkItemFile(file)
		if err != nil {
			continue
		}

		if id, err := strconv.Atoi(workItem.ID); err == nil {
			if id > maxID {
				maxID = id
			}
		}
	}

	nextID := maxID + 1
	return fmt.Sprintf("%03d", nextID), nil
}

func FixDuplicateIDs() (*ValidationResult, error) {
	result := &ValidationResult{}

	files, err := getWorkItemFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get work item files: %w", err)
	}

	// Group files by ID
	idGroups := make(map[string][]string)
	for _, file := range files {
		workItem, err := parseWorkItemFile(file)
		if err != nil {
			continue
		}
		idGroups[workItem.ID] = append(idGroups[workItem.ID], file)
	}

	// Fix duplicates by assigning new IDs to newer files
	for _, files := range idGroups {
		if len(files) > 1 {
			// Sort files by modification time (newest first)
			sort.Slice(files, func(i, j int) bool {
				info1, _ := os.Stat(files[i])
				info2, _ := os.Stat(files[j])
				return info1.ModTime().After(info2.ModTime())
			})

			// Keep the oldest file with the original ID, assign new IDs to others
			for i := 1; i < len(files); i++ {
				newID, err := GetNextID()
				if err != nil {
					result.AddError(files[i], fmt.Sprintf("failed to generate new ID: %v", err))
					continue
				}

				// Update the file with new ID
				if err := updateWorkItemID(files[i], newID); err != nil {
					result.AddError(files[i], fmt.Sprintf("failed to update ID: %v", err))
				}
			}
		}
	}

	return result, nil
}

func updateWorkItemID(filePath, newID string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Replace the ID in the YAML front matter
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "id:") {
			lines[i] = fmt.Sprintf("id: %s", newID)
			break
		}
	}

	return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
}
