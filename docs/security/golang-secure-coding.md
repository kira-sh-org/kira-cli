# Go Secure Coding Practices

This document outlines security patterns used in the kira codebase. All file operations and command executions must follow these patterns to prevent security vulnerabilities.

## File Path Validation

**Always validate file paths before use** to prevent path traversal attacks. Never use `os.ReadFile()` or `os.WriteFile()` directly with user-provided or variable paths.

### Pattern 1: Validating Paths Within a Specific Directory

For files that must be within a specific directory (e.g., `.work/`):

```go
import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

// validateWorkPath ensures a path is safe and within the .work directory
func validateWorkPath(path string) error {
    // Clean the path to remove .. and other traversal attempts
    cleanPath := filepath.Clean(path)

    // Resolve to absolute path
    absPath, err := filepath.Abs(cleanPath)
    if err != nil {
        return fmt.Errorf("invalid path: %w", err)
    }

    // Get absolute path of target directory
    workDir, err := filepath.Abs(".work")
    if err != nil {
        return fmt.Errorf("failed to resolve .work directory: %w", err)
    }

    // Ensure the path is within target directory
    // Use separator to prevent partial matches (e.g., .work-backup)
    workDirWithSep := workDir + string(filepath.Separator)
    if !strings.HasPrefix(absPath+string(filepath.Separator), workDirWithSep) && absPath != workDir {
        return fmt.Errorf("path outside .work directory: %s", path)
    }

    return nil
}

// safeReadFile reads a file after validating the path
func safeReadFile(filePath string) ([]byte, error) {
    if err := validateWorkPath(filePath); err != nil {
        return nil, err
    }
    // #nosec G304 - path has been validated by validateWorkPath above
    return os.ReadFile(filePath)
}
```

### Pattern 2: Validating Project Root Files

For files in the project root (like `RELEASES.md`, `kira.yml`):

```go
// safeReadProjectFile reads a file from project root
func safeReadProjectFile(filePath string) ([]byte, error) {
    // Clean the path to remove .. and other traversal attempts
    cleanPath := filepath.Clean(filePath)

    // Ensure it's a simple filename or relative path without traversal
    if strings.Contains(cleanPath, "..") {
        return nil, fmt.Errorf("path contains traversal: %s", filePath)
    }

    // Resolve to absolute path
    absPath, err := filepath.Abs(cleanPath)
    if err != nil {
        return nil, fmt.Errorf("invalid path: %w", err)
    }

    // Get absolute path of current directory
    currentDir, err := filepath.Abs(".")
    if err != nil {
        return nil, fmt.Errorf("failed to resolve current directory: %w", err)
    }

    // Ensure the path is within current directory
    currentDirWithSep := currentDir + string(filepath.Separator)
    if !strings.HasPrefix(absPath+string(filepath.Separator), currentDirWithSep) && absPath != currentDir {
        return nil, fmt.Errorf("path outside project directory: %s", filePath)
    }

    // #nosec G304 - path has been validated above
    return os.ReadFile(filePath)
}
```

## File Permissions

**Always use restrictive file permissions** to prevent unauthorized access:

- **Files**: Use `0o600` (read/write for owner only)
- **Directories**: Use `0o700` (read/write/execute for owner only)

```go
// Creating directories
if err := os.MkdirAll(directoryPath, 0o700); err != nil {
    return fmt.Errorf("failed to create directory: %w", err)
}

// Writing files
if err := os.WriteFile(filePath, data, 0o600); err != nil {
    return fmt.Errorf("failed to write file: %w", err)
}
```

**Rationale**: Files created by kira may contain sensitive work items. Restrictive permissions ensure only the owner can read/write them.

## Command Execution Security

**Always use `exec.CommandContext` with timeouts** and validate/sanitize input:

### Pattern 1: Git Commands with Input Validation

```go
import (
    "context"
    "fmt"
    "os/exec"
    "strings"
    "time"
)

// sanitizeCommitMessage validates and sanitizes a commit message
func sanitizeCommitMessage(msg string) (string, error) {
    // Remove newlines and other dangerous characters
    msg = strings.ReplaceAll(msg, "\n", " ")
    msg = strings.ReplaceAll(msg, "\r", "")
    msg = strings.TrimSpace(msg)

    // Validate length
    if len(msg) == 0 {
        return "", fmt.Errorf("commit message cannot be empty")
    }
    if len(msg) > 1000 {
        return "", fmt.Errorf("commit message too long (max 1000 characters)")
    }

    // Check for shell metacharacters that could be dangerous
    dangerous := []string{"`", "$", "(", ")", "{", "}", "[", "]", "|", "&", ";", "<", ">"}
    for _, char := range dangerous {
        if strings.Contains(msg, char) {
            return "", fmt.Errorf("commit message contains invalid character: %s", char)
        }
    }

    return msg, nil
}

func commitChanges(message string) error {
    // Sanitize commit message
    sanitized, err := sanitizeCommitMessage(message)
    if err != nil {
        return fmt.Errorf("invalid commit message: %w", err)
    }

    // Use context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // #nosec G204 - sanitized message has been validated and sanitized
    cmd := exec.CommandContext(ctx, "git", "commit", "-m", sanitized)
    return cmd.Run()
}
```

### Pattern 2: Build Commands in Tests

For test code that builds binaries:

```go
// validateTestPath ensures a path is within the test's temporary directory
func validateTestPath(path, tmpDir string) error {
    absPath, err := filepath.Abs(path)
    if err != nil {
        return fmt.Errorf("invalid path: %w", err)
    }

    absTmpDir, err := filepath.Abs(tmpDir)
    if err != nil {
        return fmt.Errorf("invalid tmpDir: %w", err)
    }

    tmpDirWithSep := absTmpDir + string(filepath.Separator)
    if !strings.HasPrefix(absPath+string(filepath.Separator), tmpDirWithSep) && absPath != absTmpDir {
        return fmt.Errorf("path outside test directory: %s", path)
    }

    return nil
}

func buildTestBinary(t *testing.T, tmpDir string) string {
    outPath := filepath.Join(tmpDir, "kira")

    // Validate output path
    require.NoError(t, validateTestPath(outPath, tmpDir))

    // Use context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // #nosec G204 - outPath validated above, command is hardcoded "go build"
    cmd := exec.CommandContext(ctx, "go", "build", "-o", outPath, "cmd/kira/main.go")
    // ... rest of build logic
}
```

## Key Principles

1. **Never trust user input**: Always validate paths and sanitize command arguments
2. **Use absolute paths for validation**: Convert to absolute paths before checking boundaries
3. **Prevent path traversal**: Use `filepath.Clean()` and check for `..` sequences
4. **Use context timeouts**: All `exec.Command()` calls should use `exec.CommandContext()` with timeouts
5. **Restrictive permissions**: Use `0o600` for files, `0o700` for directories
6. **Validate before use**: Validate paths before file operations, sanitize input before command execution
7. **Document exceptions**: Use `#nosec` comments only when validation has occurred, with explanation

## When to Use `#nosec` Comments

Only use `#nosec` comments when:
- Paths have been validated by a validation function
- Input has been sanitized by a sanitization function
- The comment explains WHY it's safe (e.g., "path validated by validateWorkPath above")

Never use `#nosec` to bypass security checks without proper validation.

## Examples from Codebase

- **Path validation**: See `internal/commands/utils.go` - `validateWorkPath()`, `safeReadFile()`
- **Project file validation**: See `internal/commands/utils.go` - `safeReadProjectFile()`
- **Template validation**: See `internal/templates/templates.go` - `validateTemplatePath()`
- **Command sanitization**: See `internal/commands/save.go` - `sanitizeCommitMessage()`
- **Test path validation**: See `internal/commands/integration_test.go` - `validateTestPath()`



