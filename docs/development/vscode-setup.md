# VS Code Development Setup

## Overview

This document describes the VS Code configuration for gh-issue-dependency development. The setup provides automatic formatting, linting, and integrated development workflows for Go and markdown.

## Required Extensions

The project includes a `.vscode/extensions.json` file that VS Code will automatically suggest installing:

### Core Extensions
- **golang.go** - Go language support with IntelliSense, debugging, and testing
- **DavidAnson.vscode-markdownlint** - Markdown linting for documentation
- **redhat.vscode-yaml** - YAML support for GitHub Actions workflows

### Recommended Extensions  
- **yzhang.markdown-all-in-one** - Enhanced markdown editing features
- **bierner.markdown-preview-github-styles** - GitHub-style markdown preview
- **esbenp.prettier-vscode** - JSON and general formatting
- **eamodio.gitlens** - Enhanced Git integration
- **GitHub.vscode-pull-request-github** - GitHub PR integration
- **streetsidesoftware.code-spell-checker** - Spell checking for code and docs
- **timonwong.shellcheck** - Shell script linting
- **editorconfig.editorconfig** - Editor configuration support

## Automatic Setup

When you open the project in VS Code:

1. **Extension Recommendations**: VS Code will prompt to install recommended extensions
2. **Settings Applied**: Development settings are automatically configured
3. **Tasks Available**: Press `Ctrl+Shift+P` → "Tasks: Run Task" to see available tasks

## Development Workflow

### Go Development

**Automatic Features:**
- Format on save using `goimports`
- Lint on save using `golangci-lint`
- Organize imports automatically
- Build and vet on save

**Manual Tasks:**
- `Ctrl+Shift+P` → "Go: Test All" - Run all tests
- `Ctrl+Shift+P` → "Go: Test Current Package" - Test current directory
- `Ctrl+Shift+P` → "Development: Full Check" - Complete validation pipeline

### Markdown Documentation

**Automatic Features:**
- Format on save
- Spell checking
- Rule validation (MD013, MD033, MD041, MD024 disabled)

**Manual Tasks:**
- `Ctrl+Shift+P` → "Markdown: Lint" - Check all markdown files

## Available Tasks

Access via `Ctrl+Shift+P` → "Tasks: Run Task":

### Go Tasks
- **Go: Build** - Build the project
- **Go: Test All** - Run all tests with verbose output
- **Go: Test Current Package** - Test current directory only
- **Go: Test with Coverage** - Run tests with coverage report
- **Go: Lint** - Run golangci-lint on all packages
- **Go: Format** - Format all Go files with goimports
- **Go: Mod Tidy** - Clean up go.mod and go.sum
- **Go: Vet** - Run go vet on all packages

### Development Workflow
- **Development: Full Check** - Complete validation (format → tidy → vet → lint → test)
- **Markdown: Lint** - Validate all markdown files
- **Git: Status** - Show git status
- **GitHub CLI: Create PR** - Create pull request with gh CLI
- **MkDocs: Serve** - Start documentation server

## Keyboard Shortcuts

### Default VS Code + Go Extension
- `Ctrl+Shift+P` - Command palette
- `F5` - Run/debug current file
- `Ctrl+F5` - Run without debugging  
- `Ctrl+Shift+T` - Run tests in current file
- `F12` - Go to definition
- `Shift+F12` - Find all references
- `Ctrl+.` - Quick fix / code actions

### Custom Shortcuts (Recommended)
Add to VS Code keybindings.json:

```json
[
  {
    "key": "ctrl+shift+b",
    "command": "workbench.action.tasks.runTask", 
    "args": "Development: Full Check"
  },
  {
    "key": "ctrl+shift+t",
    "command": "workbench.action.tasks.runTask",
    "args": "Go: Test All"
  }
]
```

## Configuration Details

### Go Configuration
```json
{
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint", 
  "go.lintOnSave": "package",
  "go.buildOnSave": "package",
  "go.vetOnSave": "package",
  "go.testFlags": ["-v", "-race"]
}
```

### Markdown Configuration  
```json
{
  "markdownlint.config": {
    "MD013": false,  // Line length
    "MD033": false,  // Inline HTML
    "MD041": false,  // First line heading  
    "MD024": false   // Duplicate headings
  }
}
```

### File Associations
```json
{
  "files.associations": {
    "*.md": "markdown",
    ".github/workflows/*.yml": "yaml",
    "*.gotmpl": "go"
  }
}
```

## Troubleshooting

### Go Tools Not Working

**Problem**: goimports, golangci-lint not found

**Solution**:
```bash
# Install required tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Markdown Linting Errors

**Problem**: MD013 line length errors

**Solution**: Lines are allowed to be long (disabled in config). Other MD rules are relaxed for documentation.

### Extensions Not Installing

**Problem**: VS Code doesn't suggest extensions

**Solution**: 
1. Open Command Palette (`Ctrl+Shift+P`)
2. Run "Extensions: Show Recommended Extensions"
3. Install manually if needed

### Tasks Not Appearing

**Problem**: Custom tasks not in task list

**Solution**:
1. Reload window (`Ctrl+Shift+P` → "Developer: Reload Window")
2. Verify `.vscode/tasks.json` exists
3. Check for JSON syntax errors

## Integration with CI/CD

The VS Code configuration aligns with the project's CI/CD pipeline:

- **Same linting rules** as GitHub Actions
- **Same formatting** as CI validation
- **Same test flags** as automated testing
- **Pre-commit validation** matches CI requirements

### Before Pushing
Run the "Development: Full Check" task to ensure your changes will pass CI:

1. Press `Ctrl+Shift+P`
2. Type "Tasks: Run Task"
3. Select "Development: Full Check"
4. Verify all steps pass

This runs the same validation as the CI pipeline:
- Format code with goimports
- Clean dependencies with go mod tidy
- Run go vet for common errors
- Run golangci-lint for style issues
- Run all tests with race detection

## Performance Tips

### Large Codebase
- Use "Go: Test Current Package" instead of "Go: Test All" during development
- Disable "go.buildOnSave" if builds are slow

### Resource Usage
- Background tasks (MkDocs serve) run in separate terminal panels
- Linting runs only on save, not on type
- Test coverage is manual task only

### Git Integration
- GitLens provides inline blame and history
- GitHub extension enables PR review in editor
- Built-in Git panel shows file changes

---

This setup provides a professional development environment optimized for the gh-issue-dependency project workflow.