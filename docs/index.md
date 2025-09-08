# gh-issue-dependency

A powerful GitHub CLI extension for managing issue dependencies with ease.

## What is gh-issue-dependency?

`gh-issue-dependency` is a command-line tool that extends the GitHub CLI to provide comprehensive dependency management for GitHub issues. It allows you to create, view, and remove dependency relationships between issues, helping teams better organize and track project workflows.

## Key Features

- **🔗 Dependency Management**: Create "blocks" and "blocked-by" relationships between issues
- **👁️ Visualization**: List and view dependency relationships in multiple formats
- **🛡️ Safety Features**: Dry-run mode and confirmation prompts prevent accidental changes
- **⚡ Fast & Reliable**: Built in Go with comprehensive error handling and retry logic
- **🌐 Cross-Repository**: Manage dependencies across different repositories
- **📊 Multiple Formats**: Output in plain text, JSON, or TTY-optimized formats

## Quick Example

```bash
# List dependencies for an issue
gh issue-dependency list 123

# Add a dependency relationship
gh issue-dependency add 123 --blocked-by 456

# Remove a dependency (with safety confirmation)
gh issue-dependency remove 123 --blocked-by 456

# Preview changes with dry-run
gh issue-dependency add 123 --blocks 789 --dry-run
```

## Use Cases

### Sprint Planning
Organize sprint tasks by creating dependency chains that reflect your team's workflow requirements.

### Epic Management  
Break down large epics into smaller issues with clear dependency relationships to track progress and blockers.

### Cross-Team Coordination
Manage dependencies between issues across different repositories and teams.

### Release Planning
Ensure features and bug fixes are completed in the correct order by establishing clear dependency chains.

## Getting Started

Ready to start managing your issue dependencies? Check out our [Getting Started Guide](getting-started/) for installation instructions and your first dependency management workflow.

## Commands

- [`list`](commands/list.md) - Display issue dependencies
- [`add`](commands/add.md) - Create dependency relationships  
- [`remove`](commands/remove.md) - Remove dependency relationships

## Support

- [Troubleshooting](troubleshooting/) - Common issues and solutions
- [Examples](examples/) - Real-world usage scenarios
- [GitHub Issues](https://github.com/torynet/gh-issue-dependency/issues) - Report bugs or request features