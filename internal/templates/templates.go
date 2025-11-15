package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type InputType string

const (
	InputString   InputType = "string"
	InputNumber   InputType = "number"
	InputDateTime InputType = "datetime"
)

type Input struct {
	Type        InputType
	Name        string
	Description string
	Options     []string
	DateFormat  string
}

type TemplateInput struct {
	Inputs map[string]Input
}

func ParseTemplateInputs(content string) (*TemplateInput, error) {
	inputs := make(map[string]Input)

	// Regex to match input comments: <!--input-type:variable-name:"description"-->
	re := regexp.MustCompile(`<!--input-(\w+)(?:\[([^\]]+)\])?:([^:]+):"([^"]+)"-->`)

	matches := re.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) != 5 {
			continue
		}

		inputType := match[1]
		options := match[2]
		name := match[3]
		description := match[4]

		var input Input
		input.Name = name
		input.Description = description

		switch inputType {
		case "string":
			input.Type = InputString
			if options != "" {
				input.Options = strings.Split(options, ",")
			}
		case "number":
			input.Type = InputNumber
		case "datetime":
			input.Type = InputDateTime
			if options != "" {
				input.DateFormat = options
			} else {
				input.DateFormat = "2006-01-02"
			}
		case "strings":
			input.Type = InputString
			if options != "" {
				input.Options = strings.Split(options, ",")
			}
		default:
			return nil, fmt.Errorf("unknown input type: %s", inputType)
		}

		inputs[name] = input
	}

	return &TemplateInput{Inputs: inputs}, nil
}

func ProcessTemplate(templatePath string, inputs map[string]string) (string, error) {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template: %w", err)
	}

	result := string(content)

	// Replace input placeholders with provided values
	for name, value := range inputs {
		placeholder := fmt.Sprintf("<!--input-\\w+(?:\\[[^\\]]+\\])?:%s:\"[^\"]+\"-->", name)
		re := regexp.MustCompile(placeholder)
		result = re.ReplaceAllString(result, value)
	}

	// Replace any remaining input placeholders with defaults
	result = replaceRemainingInputs(result)

	return result, nil
}

func replaceRemainingInputs(content string) string {
	// Replace string inputs with empty string
	re := regexp.MustCompile(`<!--input-string(?:\[[^\]]+\])?:([^:]+):"[^"]+"-->`)
	content = re.ReplaceAllString(content, "")

	// Replace number inputs with 0
	re = regexp.MustCompile(`<!--input-number:([^:]+):"[^"]+"-->`)
	content = re.ReplaceAllString(content, "0")

	// Replace datetime inputs with current date
	re = regexp.MustCompile(`<!--input-datetime(?:\[[^\]]+\])?:([^:]+):"[^"]+"-->`)
	content = re.ReplaceAllString(content, time.Now().Format("2006-01-02"))

	// Replace strings inputs with empty array
	re = regexp.MustCompile(`<!--input-strings(?:\[[^\]]+\])?:([^:]+):"[^"]+"-->`)
	content = re.ReplaceAllString(content, "[]")

	return content
}

func GetTemplateInputs(templatePath string) ([]Input, error) {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	templateInput, err := ParseTemplateInputs(string(content))
	if err != nil {
		return nil, err
	}

	var inputs []Input
	for _, input := range templateInput.Inputs {
		inputs = append(inputs, input)
	}

	return inputs, nil
}

func CreateDefaultTemplates(basePath string) error {
	templates := map[string]string{
		"template.prd.md":   getPRDTemplate(),
		"template.issue.md": getIssueTemplate(),
		"template.spike.md": getSpikeTemplate(),
		"template.task.md":  getTaskTemplate(),
	}

	templatesDir := filepath.Join(basePath, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	for filename, content := range templates {
		path := filepath.Join(templatesDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write template %s: %w", filename, err)
		}
	}

	return nil
}

func getPRDTemplate() string {
	return `---
id: <!--input-number:id:"Work item ID"-->
title: <!--input-string:title:"Feature title"-->
status: <!--input-string[backlog,todo,doing,review,done,released,abandoned,archived]:status:"Current status"-->
kind: prd
assigned: <!--input-string:assigned:"Assigned to (email)"-->
estimate: <!--input-number:estimate:"Estimate in days"-->
created: <!--input-datetime[yyyy-mm-dd]:created:"Creation date"-->
due: <!--input-datetime[yyyy-mm-dd]:due:"Due date (optional)"-->
tags: <!--input-strings[frontend,backend,database,api,ui,security]:tags:"Tags"-->
---

# <!--input-string:title:"Feature title"-->

## Context
<!--input-string:context:"Background and rationale"-->

## Requirements
<!--input-string:requirements:"Functional requirements"-->

## Acceptance Criteria
- [ ] <!--input-string:criteria1:"First acceptance criterion"-->
- [ ] <!--input-string:criteria2:"Second acceptance criterion"-->

## Implementation Notes
<!--input-string:implementation:"Technical implementation details"-->

## Release Notes
<!--input-string:release_notes:"Public-facing changes (optional)"-->
`
}

func getIssueTemplate() string {
	return `---
id: <!--input-number:id:"Issue ID"-->
title: <!--input-string:title:"Issue title"-->
status: <!--input-string[backlog,todo,doing,review,done,released,abandoned,archived]:status:"Current status"-->
kind: issue
assigned: <!--input-string:assigned:"Assigned to (email)"-->
estimate: <!--input-number:estimate:"Estimate in days"-->
created: <!--input-datetime[yyyy-mm-dd]:created:"Creation date"-->
tags: <!--input-strings[bug,performance,security,ui]:tags:"Tags"-->
---

# <!--input-string:title:"Issue title"-->

## Problem Description
<!--input-string:problem:"What is the problem?"-->

## Steps to Reproduce
1. <!--input-string:step1:"First step"-->
2. <!--input-string:step2:"Second step"-->
3. <!--input-string:step3:"Third step"-->

## Expected Behavior
<!--input-string:expected:"What should happen?"-->

## Actual Behavior
<!--input-string:actual:"What actually happens?"-->

## Solution
<!--input-string:solution:"Proposed solution"-->

## Release Notes
<!--input-string:release_notes:"Public-facing changes (optional)"-->
`
}

func getSpikeTemplate() string {
	return `---
id: <!--input-number:id:"Spike ID"-->
title: <!--input-string:title:"Spike title"-->
status: <!--input-string[backlog,todo,doing,review,done,released,abandoned,archived]:status:"Current status"-->
kind: spike
assigned: <!--input-string:assigned:"Assigned to (email)"-->
estimate: <!--input-number:estimate:"Estimate in days"-->
created: <!--input-datetime[yyyy-mm-dd]:created:"Creation date"-->
tags: <!--input-strings[research,discovery,investigation]:tags:"Tags"-->
---

# <!--input-string:title:"Spike title"-->

## Objective
<!--input-string:objective:"What are we trying to understand?"-->

## Questions to Answer
- <!--input-string:question1:"First question"-->
- <!--input-string:question2:"Second question"-->

## Approach
<!--input-string:approach:"How will we investigate?"-->

## Findings
<!--input-string:findings:"What did we discover?"-->

## Recommendations
<!--input-string:recommendations:"What should we do next?"-->

## Release Notes
<!--input-string:release_notes:"Public-facing changes (optional)"-->
`
}

func getTaskTemplate() string {
	return `---
id: <!--input-number:id:"Task ID"-->
title: <!--input-string:title:"Task title"-->
status: <!--input-string[backlog,todo,doing,review,done,released,abandoned,archived]:status:"Current status"-->
kind: task
assigned: <!--input-string:assigned:"Assigned to (email)"-->
estimate: <!--input-number:estimate:"Estimate in days"-->
created: <!--input-datetime[yyyy-mm-dd]:created:"Creation date"-->
tags: <!--input-strings[implementation,maintenance,refactoring]:tags:"Tags"-->
---

# <!--input-string:title:"Task title"-->

## Description
<!--input-string:description:"What needs to be done?"-->

## Steps
1. <!--input-string:step1:"First step"-->
2. <!--input-string:step2:"Second step"-->
3. <!--input-string:step3:"Third step"-->

## Definition of Done
- [ ] <!--input-string:done1:"First completion criterion"-->
- [ ] <!--input-string:done2:"Second completion criterion"-->

## Notes
<!--input-string:notes:"Additional notes"-->

## Release Notes
<!--input-string:release_notes:"Public-facing changes (optional)"-->
`
}
