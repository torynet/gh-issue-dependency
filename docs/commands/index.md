# Command Reference

Complete reference for all gh-issue-dependency commands.

## Overview

gh-issue-dependency provides three core commands for managing issue dependencies:

- **[`list`](list.md)** - Display existing dependencies for an issue
- **[`add`](add.md)** - Create new dependency relationships
- **[`remove`](remove.md)** - Remove existing dependency relationships

## Global Options

These options are available for all commands:

### `--repo <owner/repo>`
Specify the repository when not in a git repository or to work with a different repository.

```bash
gh issue-dependency list 123 --repo octocat/Hello-World
```

### `--help`
Show help information for any command.

```bash
gh issue-dependency --help
gh issue-dependency list --help
```

## Issue Reference Formats

All commands accept issues in multiple formats:

### Issue Number
```bash
gh issue-dependency list 123
```

### Repository/Issue Format
```bash
gh issue-dependency list octocat/Hello-World#123
```

### GitHub URL
```bash
gh issue-dependency list https://github.com/octocat/Hello-World/issues/123
```

## Output Formats

Most commands support different output formats:

### Default (TTY)
Optimized for terminal display with colors and formatting.

### Plain Text
Clean text output without colors, suitable for scripts.

### JSON
Machine-readable JSON output for integration with other tools.

```bash
gh issue-dependency list 123 --format json
```

## Common Patterns

### Working with Multiple Issues

Many commands support comma-separated lists:

```bash
# Add multiple dependencies at once
gh issue-dependency add 123 --blocked-by 456,789,101

# Remove multiple dependencies
gh issue-dependency remove 123 --blocked-by 456,789
```

### Cross-Repository Dependencies

Work with dependencies across different repositories:

```bash
# Add dependency from another repository
gh issue-dependency add 123 --blocked-by octocat/Hello-World#456

# List dependencies including cross-repo relationships
gh issue-dependency list 123 --repo myorg/myproject
```

### Safety and Preview

Use dry-run mode to preview changes:

```bash
# Preview what would be added
gh issue-dependency add 123 --blocks 456 --dry-run

# Preview what would be removed
gh issue-dependency remove 123 --all --dry-run
```

## Error Handling

The extension provides clear error messages and suggestions:

- **Authentication errors**: Guidance on GitHub CLI authentication
- **Permission errors**: Information about required repository access
- **Validation errors**: Clear explanations of what went wrong and how to fix it
- **API errors**: User-friendly explanations of GitHub API issues

## Next Steps

- **[`list` command](list.md)** - Learn about viewing dependencies
- **[`add` command](add.md)** - Create new dependency relationships  
- **[`remove` command](remove.md)** - Remove existing dependencies
- **[Examples](../examples/)** - See real-world usage scenarios