# gh-issue-dependency

A GitHub CLI extension for managing issue dependencies using GitHub's native dependency API. This tool helps you organize complex projects by creating and managing dependency relationships between issues, whether in the same repository or across different repositories.

## Features

- **Issue Dependencies**: Create blocking and blocked-by relationships between issues
- **Cross-Repository Support**: Manage dependencies across different repositories  
- **Multiple Output Formats**: View dependencies in table, JSON, or CSV format
- **GitHub CLI Integration**: Seamless integration with existing GitHub CLI workflows
- **Native API**: Uses GitHub's official dependency API for maximum compatibility
- **Comprehensive Validation**: Prevents circular dependencies and validates permissions

## Installation

### Prerequisites

- **Go 1.19 or later** - [Install Go](https://golang.org/doc/install)
- **GitHub CLI** - [Install GitHub CLI](https://cli.github.com/) and authenticate with `gh auth login`

### Install as GitHub CLI Extension (Recommended)

```bash
# Install the extension
gh extension install torynet/gh-issue-dependency

# Verify installation
gh issue-dependency --help
```

### Install from Source

```bash
# Clone the repository
git clone https://github.com/torynet/gh-issue-dependency
cd gh-issue-dependency

# Build the binary
go build -o gh-issue-dependency

# Move to your PATH (optional)
sudo mv gh-issue-dependency /usr/local/bin/
```

### Install via Go

```bash
go install github.com/torynet/gh-issue-dependency@latest
```

## Quick Start

### Authentication

This extension uses your existing GitHub CLI authentication. Verify you're authenticated:

```bash
gh auth status
```

If you need to authenticate:

```bash
gh auth login
```

### Basic Usage

```bash
# List all dependencies for issue #123
gh issue-dependency list 123

# Make issue #123 depend on issue #456 (123 is blocked by 456)
gh issue-dependency add 123 --blocked-by 456

# Make issue #123 block issue #789 (789 is blocked by 123)  
gh issue-dependency add 123 --blocks 789

# Remove a dependency relationship
gh issue-dependency remove 123 --blocked-by 456
```

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

## Project Structure

```
gh-issue-dependency/
├── main.go                      # Application entry point
├── cmd/                         # Command implementations
│   ├── root.go                  # Root command and global flags
│   ├── list.go                  # List command implementation
│   ├── add.go                   # Add command implementation
│   └── remove.go                # Remove command implementation
├── pkg/                         # Shared utilities and types
│   ├── github.go                # GitHub API integration
│   └── errors.go                # Error handling and formatting
├── tests/                       # Integration tests
│   └── integration_test.sh      # Test suite
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── README.md                    # This file
├── CONTRIBUTING.md              # Developer guide
└── LICENSE                      # MIT license
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on:
- Development setup
- Running tests
- Code style guidelines
- Submitting pull requests

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/torynet/gh-issue-dependency/issues)
- **Discussions**: [GitHub Discussions](https://github.com/torynet/gh-issue-dependency/discussions)  
- **Documentation**: This README and `gh issue-dependency <command> --help`

## Related Projects

- [GitHub CLI](https://cli.github.com/) - The official GitHub command line tool
- [GitHub Issues](https://docs.github.com/en/issues) - GitHub's issue tracking documentation
- [GitHub Dependencies API](https://docs.github.com/en/rest/issues/issues#create-an-issue-repository) - The underlying API this tool uses