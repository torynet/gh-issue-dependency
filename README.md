# gh-issue-dependency

[![Go Report Card](https://goreportcard.com/badge/github.com/torynet/gh-issue-dependency)](https://goreportcard.com/report/github.com/torynet/gh-issue-dependency)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GitHub Release](https://img.shields.io/github/release/torynet/gh-issue-dependency.svg)](https://github.com/torynet/gh-issue-dependency/releases)
[![Documentation](https://img.shields.io/badge/docs-gh--pages-blue)](https://torynet.github.io/gh-issue-dependency/)

A powerful GitHub CLI extension for managing issue dependencies with comprehensive validation, safety features, and cross-repository support. Organize complex projects by creating and managing dependency relationships between GitHub issues.

## ✨ Features

- 🔗 **Complete Dependency Management** - Create, view, and remove "blocks" and "blocked-by" relationships
- 🛡️ **Safety First** - Dry-run mode, confirmation prompts, and circular dependency prevention
- 🌐 **Cross-Repository Support** - Manage dependencies across different repositories and organizations
- 📊 **Multiple Output Formats** - TTY-optimized, JSON, and plain text formats for any workflow
- ⚡ **Performance & Reliability** - Built-in retry logic, rate limiting handling, and comprehensive error messages
- 🎯 **Batch Operations** - Handle multiple dependencies efficiently with comma-separated lists
- 🔐 **Enterprise Ready** - GitHub Enterprise Server support with proper authentication and permissions

## 🚀 Installation

### Prerequisites

- **GitHub CLI** - [Install GitHub CLI](https://cli.github.com/) and authenticate with `gh auth login`
- **Git** installed on your system

### GitHub CLI Extension (Recommended)

```bash
# Install as GitHub CLI extension
gh extension install torynet/gh-issue-dependency

# Verify installation
gh issue-dependency --help
```

### Alternative Installation (Development/Testing)

```bash
# Install via Go (requires Go 1.19+)
go install github.com/torynet/gh-issue-dependency@latest

# Verify installation (note: standalone binary)
gh-issue-dependency --help
```

### Download Binary

1. Visit [Releases](https://github.com/torynet/gh-issue-dependency/releases)
2. Download the binary for your system
3. Extract and place in your PATH

### Package Managers

```bash
# Homebrew (macOS/Linux)
brew install torynet/tap/gh-issue-dependency

# Chocolatey (Windows)
choco install gh-issue-dependency

# Scoop (Windows)
scoop install gh-issue-dependency
```

📖 **[Complete Installation Guide →](https://torynet.github.io/gh-issue-dependency/getting-started/)**

## 🏃 Quick Start

### 1. Authenticate with GitHub

```bash
# Check authentication status
gh auth status

# Login if needed
gh auth login
```

### 2. Your First Dependency

```bash
# Navigate to your repository
cd /path/to/your/repo

# List current dependencies
gh issue-dependency list 123

# Create a dependency (issue #123 is blocked by #456)
gh issue-dependency add 123 --blocked-by 456

# Verify the relationship was created
gh issue-dependency list 123
```

### 3. Preview Changes Safely

```bash
# Preview what would be created
gh issue-dependency add 123 --blocks 789 --dry-run

# Execute after reviewing
gh issue-dependency add 123 --blocks 789
```

🎓 **[Full Tutorial →](https://torynet.github.io/gh-issue-dependency/getting-started/)**

## Detailed Usage

### Listing Dependencies

View all dependencies for an issue:

```bash
# Basic list
gh issue-dependency list 123

# List with detailed information
gh issue-dependency list 123 --detailed

# Export to JSON for scripting
gh issue-dependency list 123 --format json

# Export to CSV for analysis
gh issue-dependency list 123 --format csv > dependencies.csv

# List dependencies for issue in another repository
gh issue-dependency list 456 --repo owner/other-repo
```

### Adding Dependencies

Create dependency relationships between issues:

```bash
# Make issue #123 depend on issue #456
gh issue-dependency add 123 --blocked-by 456

# Make issue #123 block multiple issues  
gh issue-dependency add 123 --blocks 456,789,101

# Add cross-repository dependency
gh issue-dependency add 123 --blocked-by owner/other-repo#456

# Work with issues in a specific repository
gh issue-dependency add 123 --blocks 456 --repo owner/project
```

### Removing Dependencies

Remove existing dependency relationships:

```bash
# Remove issue #456 from blocking issue #123
gh issue-dependency remove 123 --blocked-by 456

# Remove multiple blocking relationships
gh issue-dependency remove 123 --blocked-by 456,789

# Remove cross-repository dependency
gh issue-dependency remove 123 --blocked-by owner/other-repo#456
```

### Understanding Relationships

- **`--blocked-by`**: The issue cannot be completed until the specified issues are done
- **`--blocks`**: The specified issues cannot be completed until this issue is done

Example workflow:
```bash
# Issue #1 must be done before #2 can start
gh issue-dependency add 2 --blocked-by 1

# Issue #2 must be done before #3 and #4 can start  
gh issue-dependency add 2 --blocks 3,4
```

## Issue Reference Formats

Issues can be referenced in multiple ways:

- **Same repository**: `123` or `#123`
- **Cross-repository**: `owner/repo#123`
- **Multiple issues**: `123,456,789` (comma-separated, no spaces)

## Output Formats

### Table Format (Default)

Human-readable format showing issue numbers, titles, and states:

```
BLOCKING ISSUES
#456  Implement authentication  [open]
#789  Setup database schema     [closed]

BLOCKED ISSUES  
#101  Add user dashboard        [open]
#102  Create admin panel        [draft]
```

### JSON Format

Machine-readable format for scripting:

```json
{
  "issue": 123,
  "repository": "owner/repo",
  "blocking": [
    {
      "number": 456,
      "title": "Implement authentication",
      "state": "open",
      "repository": "owner/repo"
    }
  ],
  "blocked": [
    {
      "number": 101,
      "title": "Add user dashboard", 
      "state": "open",
      "repository": "owner/repo"
    }
  ]
}
```

### CSV Format

Comma-separated values for spreadsheet import:

```csv
Type,Number,Title,State,Repository
blocking,456,Implement authentication,open,owner/repo
blocked,101,Add user dashboard,open,owner/repo
```

## Troubleshooting

### Authentication Issues

**Problem**: `authentication required` error

**Solution**: 
```bash
gh auth status  # Check current authentication
gh auth login   # Authenticate if needed
```

### Permission Issues  

**Problem**: `permission denied` when modifying issues

**Solution**: Ensure you have write access to the repository. For organization repositories, you may need:
- Write permissions on the repository
- Appropriate organization role
- Issues feature enabled

### Repository Not Found

**Problem**: `repository not found` error

**Solution**:
```bash
# Specify repository explicitly
gh issue-dependency list 123 --repo owner/correct-repo

# Check repository name and access
gh repo view owner/repo
```

### Invalid Issue Numbers

**Problem**: `issue not found` or invalid reference errors

**Solution**:
- Verify issue numbers exist: `gh issue view 123`
- Check repository for cross-repository references
- Ensure proper format: `owner/repo#123` for cross-repository

### Rate Limiting

**Problem**: `rate limit exceeded` errors

**Solution**:
- Wait for rate limit to reset (usually 1 hour)
- Use authenticated requests (this extension automatically uses GitHub CLI auth)
- For GitHub Enterprise, check with your administrator

## Advanced Usage

### Scripting Integration

Use JSON output for automation:

```bash
#!/bin/bash
# Get all blocking issues as JSON
dependencies=$(gh issue-dependency list 123 --format json)

# Extract issue numbers using jq
blocking_issues=$(echo "$dependencies" | jq -r '.blocking[].number')

# Process each blocking issue
for issue in $blocking_issues; do
    echo "Checking status of issue #$issue..."
    gh issue view "$issue" --json state
done
```

### Batch Operations

Manage multiple dependencies efficiently:

```bash
# Add multiple dependencies at once
gh issue-dependency add 123 --blocked-by 1,2,3,4,5

# Remove all blocking relationships (requires listing first)
blocking=$(gh issue-dependency list 123 --format json | jq -r '.blocking[].number' | tr '\n' ',')
gh issue-dependency remove 123 --blocked-by "${blocking%,}"
```

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Development setup and testing
- Code style and guidelines  
- Pull request process

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🆘 Support

- 📖 **[Documentation](https://torynet.github.io/gh-issue-dependency/)** - Comprehensive guides and examples
- 🐛 **[Issues](https://github.com/torynet/gh-issue-dependency/issues)** - Bug reports and feature requests
- 💬 **[Discussions](https://github.com/torynet/gh-issue-dependency/discussions)** - Questions and community
- 📝 **Help**: Run `gh issue-dependency <command> --help` for command-specific help

## 🌟 Star History

If this tool helps you manage your projects better, please consider giving it a star! ⭐

---

<div align="center">
  <p><strong>Made with ❤️ for the GitHub community</strong></p>
  <p>Built with <a href="https://go.dev/">Go</a> • Powered by <a href="https://cli.github.com/">GitHub CLI</a></p>
</div>