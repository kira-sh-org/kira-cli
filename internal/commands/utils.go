package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// findWorkItemFile searches for a work item file by ID
func findWorkItemFile(workItemID string) (string, error) {
	var foundPath string

	err := filepath.Walk(".work", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if this is a work item file with the matching ID
		if strings.HasSuffix(path, ".md") && !strings.Contains(path, "template") && !strings.HasSuffix(path, "IDEAS.md") {
			// Read the file to check the ID
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Simple check for ID in front matter
			if strings.Contains(string(content), fmt.Sprintf("id: %s", workItemID)) {
				foundPath = path
				return filepath.SkipDir
			}
		}

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to search for work item: %w", err)
	}

	if foundPath == "" {
		return "", fmt.Errorf("work item with ID %s not found", workItemID)
	}

	return foundPath, nil
}

// updateWorkItemStatus updates the status field in a work item file
func updateWorkItemStatus(filePath, newStatus string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "status:") {
			lines[i] = fmt.Sprintf("status: %s", newStatus)
			break
		}
	}

	return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0o600)
}

// getWorkItemFiles returns all work item files in a directory
func getWorkItemFiles(sourcePath string) ([]string, error) {
	var files []string

	err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".md") && !strings.Contains(path, "template") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// archiveWorkItems archives work items to the archive directory
func archiveWorkItems(workItems []string, sourcePath string) (string, error) {
	// Create archive directory
	date := time.Now().Format("2006-01-02")
	archiveDir := filepath.Join(".work", "z_archive", date, filepath.Base(sourcePath))

	if err := os.MkdirAll(archiveDir, 0o700); err != nil {
		return "", fmt.Errorf("failed to create archive directory: %w", err)
	}

	// Copy work items to archive
	for _, workItem := range workItems {
		filename := filepath.Base(workItem)
		archivePath := filepath.Join(archiveDir, filename)

		content, err := os.ReadFile(workItem)
		if err != nil {
			return "", fmt.Errorf("failed to read work item: %w", err)
		}

		if err := os.WriteFile(archivePath, content, 0o600); err != nil {
			return "", fmt.Errorf("failed to write to archive: %w", err)
		}
	}

	return archiveDir, nil
}
