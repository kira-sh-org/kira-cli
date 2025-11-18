# Kira

A git-based, plaintext productivity tool designed with both clankers (LLMs) and meatbags (people) in mind.

## Overview

Kira uses a combination of plaintext markdown files, git, and a lightweight CLI to manage and coordinate work. This approach means clankers can directly read, write, and commit changes without complex APIs, while meatbags get full transparency into what the clankers are doing through git history.

Unlike Jira and other tools that are overly complicated and expensive, Kira keeps it simple. It's free and open source - Kira could be the Jira killer you've been waiting for. Just write the name of your task in the Death Note... I mean, in the markdown file, and watch it get done.

## Installation

### Download Pre-built Binaries (Recommended)

Pre-built binaries are available for Linux, macOS, and Windows on the [GitHub Releases](https://github.com/kira-os/kira-cli/releases) page.

**Linux/macOS:**
```bash
# Download and extract the appropriate archive for your platform
# Example for Linux amd64:
wget https://github.com/kira-os/kira-cli/releases/latest/download/kira_linux_amd64.tar.gz
tar -xzf kira_linux_amd64.tar.gz
sudo mv kira /usr/local/bin/

# Or for macOS arm64:
wget https://github.com/kira-os/kira-cli/releases/latest/download/kira_darwin_arm64.tar.gz
tar -xzf kira_darwin_arm64.tar.gz
sudo mv kira /usr/local/bin/
```

**Windows:**
```powershell
# Download and extract the zip file
# Extract kira.exe and add it to your PATH
```

### Using Go Install

```bash
go install github.com/kira-os/kira-cli/cmd/kira@latest
```

### Build from Source (Contributors)

If you want to build and test locally, see the contributor guide:

```bash
# Development setup, building, testing, and install options:
see CONTRIBUTING.md
```

## Quick Start

1. **Initialize a workspace:**
   ```bash
   kira init
   ```

2. **Create a new work item:**
   ```bash
   # Status is optional; explicit example with status before title
   kira new prd todo "User Authentication" "Implement user login system"
   ```

3. **Move work items:**
   ```bash
   kira move 001 doing
   ```

4. **Save and commit changes:**
   ```bash
   kira save "Add authentication requirements"
   ```

## Commands

### `kira init [folder]`
Creates the files and folders used by kira in the specified directory. If a `.work/` directory already exists, you can choose how to proceed using flags or interactively.

```bash
kira init                              # Initialize in current directory
kira init ~/my-project                # Initialize in specific directory
kira init --fill-missing              # Add any missing files/folders, keep existing
kira init --force                     # Overwrite existing .work (fresh init)
```

Notes:
- Creates status folders and template files.
- Adds `.gitkeep` files to empty folders.
- Without flags, if `.work/` exists you'll be prompted to cancel, overwrite, or fill-missing.

### `kira new [template] [status] [title] [description]`
Creates a new work item from a template.

```bash
kira new                                              # Prompts for template selection
kira new prd doing "Fix login bug"                    # Status before title
kira new prd "Feature"                                # Status omitted → defaults to backlog
kira new prd backlog "Feature"                        # Explicit status
kira new prd "Feature" --interactive                  # Enable prompts for missing fields
kira new prd "Feature" -I                            # Shorthand for --interactive
kira new prd "Feature" --input due=2025-01-01        # Provide inputs (key=value)
kira new prd "Feature" --input assigned=me@acme.com  # Multiple --input allowed
```

Notes:
- By default, only provided values are filled; missing template fields use defaults
- Use `--interactive` (or `-I`) to enable prompts for missing template fields

### `kira move <work-item-id> [target-status]`
Moves a work item to a different status folder.

```bash
kira move 001              # Show status options
kira move 001 doing        # Move to doing folder
```

### `kira idea <description>`
Adds an idea to the IDEAS.md file.

```bash
kira idea "Add dark mode support"
```

### `kira lint`
Scans for issues in work items.

```bash
kira lint
```

### `kira doctor`
Checks for and fixes duplicate work item IDs.

```bash
kira doctor
```

### `kira release [status|path] [subfolder]`
Generates release notes and archives completed work items.

```bash
kira release                    # Release from done folder
kira release done v2           # Release from done/v2 subfolder
kira release 4_done/v2         # Release from specific path
```

Behavior:
- Updates work item status to "released" before archival
- Archives to `.work/z_archive/{date}/{original-path}/`
- Prepends release notes to the configured `release.releases_file` (default `RELEASES.md`)
- Only items with a `# Release Notes` section are included in notes

### `kira abandon <work-item-id|path> [reason|subfolder]`
Archives work items and marks them as abandoned.

```bash
kira abandon 001                                    # Abandon single item
kira abandon 001 "Superseded by new approach"       # With reason
kira abandon done v2                                # Abandon folder
```

Behavior:
- Updates work item status to "abandoned" and archives the item(s)
- Archives to `.work/z_archive/{date}/{id}/` or `.work/z_archive/{date}/{original-path}/`
- Preserves folder structure for path/subfolder abandons
- Adds an "Abandonment" section with reason and timestamp when a reason is provided

### `kira save [commit-message]`
Updates work items and commits changes to git.

```bash
kira save                           # Use default message
kira save "Add user auth requirements"  # Custom message
```

Behavior:
- Validates all non-archived work items before staging; fails on validation errors
- Updates/creates the `updated:` timestamp in changed work items
- Stages only `.work/` changes; skips committing if external (non-.work) changes are detected
- Uses provided commit message or the configured default when none is given

### `kira version`
Prints version information embedded at build time (SemVer tag if present), commit, build date, and dirty state.

```bash
kira version
# Version: v0.1.0
# Commit: abc1234
# BuildDate: 2025-01-01T00:00:00Z
# State: clean
```

## Folder Structure

```
kira.yml          # Configuration
.work/
├── 0_backlog/    # Ideas being shaped
├── 1_todo/       # Ready to work on
├── 2_doing/      # Currently in progress (one item only)
├── 3_review/     # Ready for review
├── 4_done/       # Completed work
├── templates/    # Work item templates
├── z_archive/    # Archived items
└── IDEAS.md      # Quick idea capture
```

## Work Item Types

- **PRD** (Product Requirements Document): Feature specifications
- **Issue**: Bug reports and problems
- **Spike**: Discovery and research tasks
- **Task**: Discrete implementation tasks

## Configuration

The `kira.yml` file controls the tool's behavior:

```yaml
version: "1.0"

templates:
  prd: "templates/template.prd.md"
  issue: "templates/template.issue.md"
  spike: "templates/template.spike.md"
  task: "templates/template.task.md"

status_folders:
  backlog: "0_backlog"
  todo: "1_todo"
  doing: "2_doing"
  review: "3_review"
  done: "4_done"
  archived: "z_archive"

  # Default status used when not specified in `kira new`
  default_status: "backlog"

validation:
  required_fields: ["id", "title", "status", "kind", "created"]
  id_format: "^\\d{3}$"
  status_values: ["backlog", "todo", "doing", "review", "done", "released", "abandoned", "archived"]

commit:
  default_message: "Update work items"

release:
  releases_file: "RELEASES.md"
  archive_date_format: "2006-01-02"
```

## Work Item Format

Work items are markdown files with YAML front matter:

```yaml
---
id: 001
title: User Authentication Feature
status: todo
kind: prd
assigned: user@example.com
estimate: 3 days
created: 2024-01-15T10:00:00Z
updated: 2024-01-16T14:30:00Z
due: 2024-01-20T17:00:00Z
tags: [auth, security, frontend]
---

# User Authentication Feature

## Context
Background and rationale...

## Requirements
Functional requirements...

## Acceptance Criteria
- [ ] User can log in with email/password
- [ ] User can log out
- [ ] Session is maintained across page refreshes
```

## Git Integration

Kira is designed to work seamlessly with git:

- All work items are tracked in git
- The `kira save` command commits only `.work/` changes
- External changes are detected and prevent accidental commits
- Full transparency through git history

## License

MIT License - see LICENSE file for details.

## Contributing

See CONTRIBUTING.md for the full contributor workflow, development setup, and build/install details.

