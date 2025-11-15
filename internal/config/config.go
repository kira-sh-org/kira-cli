package config

import (
	"fmt"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	Version       string            `yaml:"version"`
	Templates     map[string]string `yaml:"templates"`
	StatusFolders map[string]string `yaml:"status_folders"`
	Validation    ValidationConfig  `yaml:"validation"`
	Commit        CommitConfig      `yaml:"commit"`
	Release       ReleaseConfig     `yaml:"release"`
	DefaultStatus string            `yaml:"default_status"`
}

type ValidationConfig struct {
	RequiredFields []string `yaml:"required_fields"`
	IDFormat       string   `yaml:"id_format"`
	StatusValues   []string `yaml:"status_values"`
}

type CommitConfig struct {
	DefaultMessage string `yaml:"default_message"`
}

type ReleaseConfig struct {
	ReleasesFile      string `yaml:"releases_file"`
	ArchiveDateFormat string `yaml:"archive_date_format"`
}

var DefaultConfig = Config{
	Version: "1.0",
	Templates: map[string]string{
		"prd":   "templates/template.prd.md",
		"issue": "templates/template.issue.md",
		"spike": "templates/template.spike.md",
		"task":  "templates/template.task.md",
	},
	StatusFolders: map[string]string{
		"backlog":  "0_backlog",
		"todo":     "1_todo",
		"doing":    "2_doing",
		"review":   "3_review",
		"done":     "4_done",
		"archived": "z_archive",
	},
	DefaultStatus: "backlog",
	Validation: ValidationConfig{
		RequiredFields: []string{"id", "title", "status", "kind", "created"},
		IDFormat:       "^\\d{3}$",
		StatusValues:   []string{"backlog", "todo", "doing", "review", "done", "released", "abandoned", "archived"},
	},
	Commit: CommitConfig{
		DefaultMessage: "Update work items",
	},
	Release: ReleaseConfig{
		ReleasesFile:      "RELEASES.md",
		ArchiveDateFormat: "2006-01-02",
	},
}

func LoadConfig() (*Config, error) {
	// Prefer root-level kira.yml; fall back to legacy .work/kira.yml if present
	rootPath := "kira.yml"
	legacyPath := filepath.Join(".work", "kira.yml")

	configPath := ""
	if _, err := os.Stat(rootPath); err == nil {
		configPath = rootPath
	} else if _, err := os.Stat(legacyPath); err == nil {
		configPath = legacyPath
	} else {
		return &DefaultConfig, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge with defaults for missing fields
	mergeWithDefaults(&config)

	return &config, nil
}

func mergeWithDefaults(config *Config) {
	if config.Templates == nil {
		config.Templates = make(map[string]string)
	}
	for k, v := range DefaultConfig.Templates {
		if _, exists := config.Templates[k]; !exists {
			config.Templates[k] = v
		}
	}

	if config.StatusFolders == nil {
		config.StatusFolders = make(map[string]string)
	}
	for k, v := range DefaultConfig.StatusFolders {
		if _, exists := config.StatusFolders[k]; !exists {
			config.StatusFolders[k] = v
		}
	}

	if config.Validation.RequiredFields == nil {
		config.Validation.RequiredFields = DefaultConfig.Validation.RequiredFields
	}
	if config.Validation.IDFormat == "" {
		config.Validation.IDFormat = DefaultConfig.Validation.IDFormat
	}
	if config.Validation.StatusValues == nil {
		config.Validation.StatusValues = DefaultConfig.Validation.StatusValues
	}

	if config.Commit.DefaultMessage == "" {
		config.Commit.DefaultMessage = DefaultConfig.Commit.DefaultMessage
	}

	if config.Release.ReleasesFile == "" {
		config.Release.ReleasesFile = DefaultConfig.Release.ReleasesFile
	}
	if config.Release.ArchiveDateFormat == "" {
		config.Release.ArchiveDateFormat = DefaultConfig.Release.ArchiveDateFormat
	}

	if config.DefaultStatus == "" {
		config.DefaultStatus = DefaultConfig.DefaultStatus
	}
}

func SaveConfig(config *Config) error {
	return SaveConfigToDir(config, ".")
}

// SaveConfigToDir saves the config to the specified target directory under .work/kira.yml
func SaveConfigToDir(config *Config, targetDir string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to root-level kira.yml in the target directory
	configPath := filepath.Join(targetDir, "kira.yml")
	// Ensure targetDir exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to ensure target directory: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
