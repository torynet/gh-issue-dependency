# Getting Started

Welcome to gh-issue-dependency! This guide will get you up and running with issue dependency management in just a few minutes.

## Prerequisites

Before installing gh-issue-dependency, make sure you have:

- **GitHub CLI (gh)** installed and authenticated
  - Install: [GitHub CLI Installation Guide](https://cli.github.com/manual/installation)
  - Authenticate: `gh auth login`
- **Git** installed on your system
- **Go 1.19 or later** (only if building from source)

## Installation

### Option 1: Go Install (Recommended)

```bash
go install github.com/torynet/gh-issue-dependency@latest
```

### Option 2: Download Binary

1. Visit the [Releases page](https://github.com/torynet/gh-issue-dependency/releases)
2. Download the appropriate binary for your system
3. Extract and place in your PATH

### Option 3: Package Managers

```bash
# Homebrew (macOS/Linux)
brew install torynet/tap/gh-issue-dependency

# Chocolatey (Windows)  
choco install gh-issue-dependency

# Scoop (Windows)
scoop install gh-issue-dependency
```

## Verify Installation

After installation, verify the extension is working:

```bash
gh issue-dependency --help
```

You should see the help output with available commands.

## Authentication Setup

The extension uses your existing GitHub CLI authentication. Ensure you're authenticated and have access to the repositories you want to manage:

```bash
# Check authentication status
gh auth status

# Login if needed
gh auth login
```

### Required Permissions

Your GitHub token needs these permissions:
- **Repository access**: Read/write access to issues
- **Metadata**: Read repository metadata
- **Issues**: Read and write issues

## Quick Start Tutorial

Let's create your first dependency relationship!

### Step 1: Navigate to Your Repository

```bash
cd /path/to/your/repository
```

### Step 2: List Existing Dependencies

```bash
# List dependencies for issue #1
gh issue-dependency list 1
```

If this is your first time, you'll likely see "No dependencies found."

### Step 3: Create a Dependency

Let's say issue #2 is blocked by issue #1:

```bash
# Add a "blocked-by" relationship
gh issue-dependency add 2 --blocked-by 1
```

### Step 4: Verify the Relationship

```bash
# Check the dependency was created
gh issue-dependency list 2
```

You should see output like:
```
Issue #2: Your Issue Title
BLOCKED BY #1 - Another Issue Title
```

### Step 5: Try Dry-Run Mode

Before making changes, you can preview them:

```bash
# Preview what would happen
gh issue-dependency add 3 --blocks 2 --dry-run
```

This shows you exactly what would be created without making changes.

## Next Steps

🎉 **Congratulations!** You've successfully created your first issue dependency.

### Learn More

- **[Command Reference](../commands/)** - Detailed documentation for all commands
- **[Examples](../examples/)** - Real-world usage scenarios
- **[Troubleshooting](../troubleshooting/)** - Solutions to common issues

### Common Workflows

- **Sprint Planning**: Use dependencies to organize tasks in logical order
- **Epic Breakdown**: Create dependency chains for large features
- **Release Coordination**: Ensure features are completed in the right sequence

## Repository Context

The extension automatically detects your repository context from:
1. Current directory's git remote
2. `--repo` flag: `gh issue-dependency list 1 --repo owner/repo`
3. GitHub URLs: `gh issue-dependency list https://github.com/owner/repo/issues/1`

## Getting Help

If you encounter issues:
1. Check the [Troubleshooting Guide](../troubleshooting/)
2. Search [existing issues](https://github.com/torynet/gh-issue-dependency/issues)
3. [Open a new issue](https://github.com/torynet/gh-issue-dependency/issues/new) if needed

Ready to dive deeper? Explore the [Command Reference](../commands/) for complete documentation of all available features.