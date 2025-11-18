package templates

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTemplateInputs(t *testing.T) {
	t.Run("parses string input", func(t *testing.T) {
		content := `<!--input-string:title:"Feature title"-->`

		inputs, err := ParseTemplateInputs(content)
		require.NoError(t, err)

		assert.Len(t, inputs.Inputs, 1)
		assert.Equal(t, InputString, inputs.Inputs["title"].Type)
		assert.Equal(t, "Feature title", inputs.Inputs["title"].Description)
	})

	t.Run("parses number input", func(t *testing.T) {
		content := `<!--input-number:estimate:"Estimate in days"-->`

		inputs, err := ParseTemplateInputs(content)
		require.NoError(t, err)

		assert.Len(t, inputs.Inputs, 1)
		assert.Equal(t, InputNumber, inputs.Inputs["estimate"].Type)
	})

	t.Run("parses datetime input", func(t *testing.T) {
		content := `<!--input-datetime[yyyy-mm-dd]:created:"Creation date"-->`

		inputs, err := ParseTemplateInputs(content)
		require.NoError(t, err)

		assert.Len(t, inputs.Inputs, 1)
		assert.Equal(t, InputDateTime, inputs.Inputs["created"].Type)
		assert.Equal(t, "yyyy-mm-dd", inputs.Inputs["created"].DateFormat)
	})

	t.Run("parses string with options", func(t *testing.T) {
		content := `<!--input-string[backlog,todo,doing]:status:"Current status"-->`

		inputs, err := ParseTemplateInputs(content)
		require.NoError(t, err)

		assert.Len(t, inputs.Inputs, 1)
		assert.Equal(t, InputString, inputs.Inputs["status"].Type)
		assert.Equal(t, []string{"backlog", "todo", "doing"}, inputs.Inputs["status"].Options)
	})
}

func TestProcessTemplate(t *testing.T) {
	t.Run("replaces input placeholders", func(t *testing.T) {
		// Create a temporary template file
		templateContent := `---
id: <!--input-number:id:"Work item ID"-->
title: <!--input-string:title:"Feature title"-->
---

# <!--input-string:title:"Feature title"-->

## Context
<!--input-string:context:"Background and rationale"-->
`

		tmpFile, err := os.CreateTemp("", "template_*.md")
		require.NoError(t, err)
		defer func() { _ = os.Remove(tmpFile.Name()) }()

		require.NoError(t, os.WriteFile(tmpFile.Name(), []byte(templateContent), 0o644))

		inputs := map[string]string{
			"id":      "001",
			"title":   "Test Feature",
			"context": "This is a test feature",
		}

		result, err := ProcessTemplate(tmpFile.Name(), inputs)
		require.NoError(t, err)

		assert.Contains(t, result, "id: 001")
		assert.Contains(t, result, "title: Test Feature")
		assert.Contains(t, result, "# Test Feature")
		assert.Contains(t, result, "This is a test feature")
	})
}

func TestCreateDefaultTemplates(t *testing.T) {
	t.Run("creates default templates", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := CreateDefaultTemplates(tmpDir)
		require.NoError(t, err)

		// Check that template files were created
		templates := []string{
			"template.prd.md",
			"template.issue.md",
			"template.spike.md",
			"template.task.md",
		}

		for _, template := range templates {
			path := tmpDir + "/templates/" + template
			_, err := os.Stat(path)
			assert.NoError(t, err, "Template %s should exist", template)
		}
	})
}
