---
id: 001
title: Kira MVP
status: review
kind: prd
assigned: wkallan1984@gmail.com
estimate: 1 day
created: 2024-06-11T15:00:00Z
---
# Kira MVP

This PRD outlines the initial MVP for Kira, a git-based, plaintext productivity tool designed with both clankers (LLMs) and meatbags (people) in mind.

## Context

Agentic workflows have changed the way we develop code, no longer do we work with only other meatbags, we have clankers in the mix, working in our IDEs or off in the background potentially even coordinating with other clankers. Kira uses a combination of plaintext markdown files, git and a lightweight CLI to manage and coordinate work.

This approach means clankers can directly read, write, and commit changes without complex APIs, while meatbags get full transparency into what the clankers are doing through git history. Unlike Jira and other tools that are overly complicated and expensive, Kira keeps it simple. It's free and open source - Kira could be the Jira killer you've been waiting for. Just write the name of your task in the Death Note... I mean, in the markdown file, and watch it get done.

## Technology Stack

The CLI will be built with Go and Cobra for command-line interface management, ensuring fast performance and cross-platform compatibility. Aim for high test coverage with unit tests.


## Folder Structure

Below is the default folder structure for the tool, specified in the `kira.yml` file. Changes to this structure can be made by modifying the `kira.yml` file.

This flat structure provides clear separation between different work states while keeping everything organized and git-trackable.

```
.work/
├── 0_backlog/
├── 1_todo/
├── 2_doing/
├── 3_review/
├── 4_done/
├── templates/
├── z_archive/
└── IDEAS.md
kira.yml
```


### Files
**kira.yml**: Configuration file for managing the kira CLI tool settings and folder structure.

**IDEAS.md**: Quick capture file for unstructured ideas - a low-effort dumping ground for capturing details without the distraction of fitting them into a formal template.


### Folders
**0_backlog**: Ideas currently being shaped. Once ready, they can be moved to the todo folder.

**1_todo**: Items ready to be worked on by a meatbag or clanker.

**2_doing**: Items being worked on in the current session. Ideally, only have one item in this folder at a time, and it should only be checked in for WIP commits.

**3_review**: Add work items to this folder when ready to raise a pull request or do a shoulder check by another meatbag or clanker.

**4_done**: Move items here when complete - either as the final commit of a PR, in person, or as a separate commit after the code has been deployed.

**z_archive**: Items that have been discarded or moved out of the done folder.

**templates**: Templates for different work item types.

### Commands

#### `kira init [folder]`
Creates the files and folders used by kira in the specified directory.
- **Arguments**: `folder` (optional) - Directory to initialize, defaults to current directory
- **Example**: `kira init` or `kira init ~/my-project`

#### `kira new [template] [work-item] [status] [description]`
Creates a new work-item from a template in the specified status folder.
- **Arguments**: All optional - will prompt for selection if not provided
  - `template` - Template type (prd, issue, spike, task)
  - `title` - Name of the work item
  - `status` - Target status (backlog, todo, doing, review, done, archived)
  - `description` - Brief description
- **Flags**:
  - `--ignore-input`, `-ii` - Skip interactive input prompts
  - `--input variable-name:value` - Provide input values directly
  - `--help-inputs` - List available input variables for a template
- **Example**: `kira new` or `kira new prd "Fix login bug" todo "Authentication fails on mobile"`

#### `kira move <work-item-id> [target-status]`
Moves the work item to the target status folder.
- **Arguments**:
  - `work-item-id` (required) - ID of the work item to move
  - `target-status` (optional) - Status name (backlog, todo, doing, review, done, archived). Will display options if not provided
- **Example**: `kira move 001` or `kira move 001 doing`

#### `kira idea <description>`
Adds an idea with a timestamp to the IDEAS.md file.
- **Arguments**: `description` (required) - The idea to capture
- **Example**: `kira idea "Add dark mode support"`

#### `kira lint`
Scans folders and files to check for issues and reports any found.
- **Example**: `kira lint`

#### `kira doctor`
Checks for and fixes duplicate work item IDs by updating the latest one with a new ID.
- **Example**: `kira doctor`

#### `kira release [status|path] [subfolder]`
Generates release notes and archives completed work items.
- **Arguments**:
  - `status|path` (optional) - Status name (backlog, todo, doing, review, done, archived) or direct path to folder, defaults to done
  - `subfolder` (optional) - Subfolder within the status folder to release
- **Behavior**:
  - Generates release notes from work items with "# Release notes" sections
  - Prepends release notes to RELEASES.md file (path specified in kira.yml)
  - Updates work item status to "released" before archival
  - Archives all contents to `z_archive/{date}/{original-path}/`
  - Preserves folder structure during archival
  - Removes original files after successful archival
- **Example**: `kira release`, `kira release done`, `kira release done v2`, or `kira release 4_done/v2`

#### `kira abandon <work-item-id|path> [reason|subfolder]`
Archives work items and marks them as abandoned.
- **Arguments**:
  - `work-item-id|path` (required) - ID of specific work item or path to folder containing items to abandon
  - `reason|subfolder` (optional) - Reason for abandonment (when using ID) or subfolder within path
- **Behavior**:
  - Updates work item status to "abandoned"
  - Archives work items to `z_archive/{date}/{work-item-id}/` or `z_archive/{date}/{original-path}/`
  - Adds abandonment reason to work items if provided
  - Preserves folder structure during archival
  - Removes original files after successful archival
- **Example**: `kira abandon 001`, `kira abandon 001 "Superseded by new approach"`, `kira abandon done v2`, or `kira abandon 1_todo "Requirements changed"`

#### `kira save [commit-message]`
Updates the `updated` field in work items and commits changes to git.
- **Arguments**: `commit-message` (optional) - Custom commit message, defaults to config value
- **Behavior**:
  - Updates `updated` timestamp for all modified work items (excluding archived items)
  - Validates all non-archived work items before staging
  - Stages only changes within the `.work/` directory
  - Commits with provided message or `kira.yml` config if no external changes are staged
  - Warns and skips commit if external changes are staged
  - Fails if validation errors are found
- **Example**: `kira save` or `kira save "Add user authentication requirements"`



## Templates

Templates are markdown files with the work item type before the file extension (e.g., `template.prd.md` or `template.task.md`).

### Default Templates

Default templates are created when `kira init` is run:

| Name              | Description |
|-------------------|-------------|
| template.prd.md   | Product requirements docs for capturing details of a feature |
| template.issue.md | Details about a bug or problem |
| template.spike.md | Discovery task to understand the scope of an idea |
| template.task.md  | Discrete task that requires more details than a checklist item in a PRD |

### Template Input System

Templates support interactive input through HTML comments with the format:
`<!--input-type:variable-name:"description"-->`

#### Supported Input Types
- `string` - Free text input
- `number` - Numeric input
- `string[option1, option2, option3]` - Single selection from options
- `strings[option1, option2, option3]` - Multiple selection from options
- `datetime[date-format]` - Date input (default format: yyyy-mm-dd)

#### Input Handling Examples
- Skip prompts: `kira new prd foobar --ignore-input`
- Provide values: `kira new prd foobar --input due:"2025-10-01" --input estimate:3`
- Handle spaces: `kira new prd foobar --input description:"Feature with spaces"`
- List inputs: `kira new prd foobar --help-inputs`

## File Format Specifications

### Work Item File Naming Convention
Work items follow the pattern: `{id}-{kebab-case-name}.{template-type}.md`

Examples:
- `001-user-authentication.prd.md`
- `002-login-bug-fix.issue.md`
- `003-api-discovery.spike.md`
- `004-setup-ci-cd.task.md`

### Front Matter Structure
Each work item file must start with YAML front matter:

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
```

**Required fields**: `id`, `title`, `status`, `kind`, `created`
**Optional fields**: `assigned`, `estimate`, `updated`, `due`, `tags`

### Work Item ID Generation
- IDs are 3-digit zero-padded numbers (001, 002, 003, etc.)
- Generated sequentially based on existing items in the repository
- `kira doctor` command handles duplicate ID resolution

### File Content Structure
After front matter, files contain markdown content with these standard sections:
- `# {Title}` - Work item title
- `## Context` - Background and rationale
- `## Requirements` - Functional requirements (for PRDs)
- `## Acceptance Criteria` - Testable conditions
- `## Implementation Notes` - Technical details
- `## Release Notes` - Public-facing changes (optional)

## Configuration Schema

### kira.yml Structure
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

## Template Content Specifications

### template.prd.md
```markdown
---
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
```

### template.issue.md
```markdown
---
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
```

## Error Handling & Validation

### Lint Rules (`kira lint`)
1. **File Structure Validation**
   - Valid YAML front matter
   - Required fields present
   - Proper file naming convention
   - Valid ID format (3-digit number)

2. **Content Validation**
   - No duplicate IDs across all files
   - Valid status values
   - Valid date formats
   - Template input syntax validation

3. **Workflow Validation**
   - Only one item in `2_doing` folder
   - Items in `4_done` have completion dates
   - Referenced templates exist

### Error Messages
- `Error: Duplicate ID found: 001 in files x.prd.md and y.issue.md`
- `Error: Invalid status 'invalid' in file 001-feature.prd.md. Valid values: backlog, todo, doing, review, done`
- `Error: Multiple items in doing folder. Only one item allowed at a time.`

### `kira doctor` Functionality
- Scans for duplicate IDs
- Updates newer files with new sequential IDs
- Reports and fixes common issues
- Validates and repairs corrupted front matter

## Implementation Details

### Project Structure
```
kira/
├── cmd/
│   └── kira/
│       └── main.go
├── internal/
│   ├── commands/
│   │   ├── init.go
│   │   ├── new.go
│   │   ├── move.go
│   │   ├── idea.go
│   │   ├── lint.go
│   │   ├── doctor.go
│   │   └── release_notes.go
│   ├── config/
│   │   └── config.go
│   ├── templates/
│   │   └── templates.go
│   └── validation/
│       └── validator.go
├── templates/
│   ├── template.prd.md
│   ├── template.issue.md
│   ├── template.spike.md
│   └── template.task.md
├── go.mod
├── go.sum
└── README.md
```

### Dependencies
- **Go**: 1.21+
- **Cobra**: v1.7.0+ (CLI framework)
- **YAML**: v3.0.0+ (YAML parsing)
- **Testify**: v1.8.4+ (testing framework)

### Build Requirements
- Cross-platform binaries (Linux, macOS, Windows)
- Single binary distribution
- No external dependencies at runtime

## Acceptance Criteria

### Core Functionality
1. **Initialization**
   - [x] `kira init` creates folder structure and default templates
   - [x] `kira init <folder>` works with custom directory
   - [x] Existing files are preserved during init

2. **Work Item Creation**
   - [x] `kira new` prompts for all required information
   - [x] `kira new <template> <status> <title> <description>` creates item directly
   - [x] Template inputs are properly replaced
   - [x] File naming convention is followed

3. **Work Item Management**
   - [x] `kira move <id> <folder>` moves items between folders
   - [x] `kira move <id>` shows folder selection prompt
   - [x] Move operation updates git history

4. **Ideas Capture**
   - [x] `kira idea <description>` appends to IDEAS.md with timestamp
   - [x] Multiple ideas are properly formatted

5. **Validation & Maintenance**
   - [x] `kira lint` reports all validation issues
   - [x] `kira doctor` fixes duplicate IDs and common issues
   - [x] All commands handle missing .work directory gracefully

6. **Release Command**
   - [x] `kira release` generates notes from done folder and archives contents
   - [x] `kira release <status>` works with custom status folders
   - [x] `kira release <status> <subfolder>` works with subfolders
   - [x] `kira release <path>` works with direct folder paths
   - [x] Updates work item status to "released" before archival
   - [x] Prepends release notes to RELEASES.md file
   - [x] Archives contents to `z_archive/{date}/{original-path}/`
   - [x] Preserves folder structure during archival
   - [x] Only items with "# Release Notes" section are included in notes

7. **Abandon Command**
   - [x] `kira abandon <id>` updates status to "abandoned" and archives single work item
   - [x] `kira abandon <id> <reason>` adds abandonment reason to single work item
   - [x] `kira abandon <path>` abandons all work items in specified folder
   - [x] `kira abandon <path> <subfolder>` abandons items in subfolder within path
   - [x] Archives to `z_archive/{date}/{id}/` or `z_archive/{date}/{original-path}/` structure
   - [x] Preserves folder structure during archival
   - [x] Removes original files after successful archival

8. **Save Command**
   - [x] `kira save` updates `updated` timestamp for modified work items
   - [x] `kira save <message>` uses custom commit message
   - [x] Validates all non-archived work items before staging
   - [x] Stages only `.work/` directory changes
   - [x] Commits with provided message or config default when no external changes staged
   - [x] Warns and skips commit when external changes outside `.work/` are staged
   - [x] Fails if validation errors are found

### Performance Requirements
- Commands complete within 2 seconds for repositories with 1000+ work items
- Memory usage under 50MB for typical operations
- Binary size under 10MB

### Cross-Platform Compatibility
- Works on Linux, macOS, and Windows
- Handles different line ending formats
- Respects system file permissions

# Follow-up tasks
- [x] Ensure there's a .gitkeep file in each of the folders created by the init command.
- [x] When kira init is run, if it detects a .work directory, it should offer the user the option cancel, overwrite or create missing files and folders
- [x] status should be optional in `kira new [template] [work-item] [status] [description]` and default to a config value which should be backlog
- [x] The kira.yaml file should be exist at the same level as the .work directory - currently it is in the .work directory so let's update code to support this.
- [x] Add a script to install the kira binary globally for the user in a way that when they run the script again it will update with the latest version from the local source code.
- [x] Add a version command to the kira binary that will print the version of the kira binary along with the git hash of the code that was used to build it.

