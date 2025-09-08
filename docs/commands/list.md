# list command

Display dependency relationships for GitHub issues.

## Synopsis

```bash
gh issue-dependency list <issue> [flags]
```

## Description

The `list` command shows all dependency relationships for a specified issue. It displays both "blocks" and "blocked-by" relationships, providing a complete view of how an issue relates to others in your project.

## Usage

### Basic Usage

```bash
# List dependencies for issue #123
gh issue-dependency list 123

# List dependencies using GitHub URL
gh issue-dependency list https://github.com/owner/repo/issues/123

# List dependencies for issue in different repository
gh issue-dependency list owner/repo#123
```

### With Repository Flag

```bash
# Specify repository explicitly
gh issue-dependency list 123 --repo owner/repository
```

## Output Formats

### Default (TTY) Output

```
Issue #123: Implement user authentication system

BLOCKED BY:
  #45 - Set up database schema
  #67 - Create user model structure

BLOCKS:
  #156 - Add login form validation
  #178 - Implement password reset flow

Dependencies: 2 blocked-by, 2 blocks
```

### Plain Text Output

```bash
gh issue-dependency list 123 --format plain
```

```
Issue #123: Implement user authentication system
BLOCKED BY #45: Set up database schema
BLOCKED BY #67: Create user model structure  
BLOCKS #156: Add login form validation
BLOCKS #178: Implement password reset flow
```

### JSON Output

```bash
gh issue-dependency list 123 --format json
```

```json
{
  "issue": {
    "number": 123,
    "title": "Implement user authentication system",
    "repository": {
      "owner": "myorg",
      "name": "myproject"
    }
  },
  "blocked_by": [
    {
      "number": 45,
      "title": "Set up database schema",
      "repository": {"owner": "myorg", "name": "myproject"}
    },
    {
      "number": 67, 
      "title": "Create user model structure",
      "repository": {"owner": "myorg", "name": "myproject"}
    }
  ],
  "blocks": [
    {
      "number": 156,
      "title": "Add login form validation", 
      "repository": {"owner": "myorg", "name": "myproject"}
    },
    {
      "number": 178,
      "title": "Implement password reset flow",
      "repository": {"owner": "myorg", "name": "myproject"}
    }
  ]
}
```

## Flags

### `--format <format>`
Output format: `tty` (default), `plain`, or `json`.

### `--repo <owner/repo>`
Repository to use when not in a git repository.

### `--help`
Show help for the list command.

## Examples

### Sprint Planning

List dependencies for all issues in a milestone to understand task order:

```bash
# Check dependencies for key sprint issues
gh issue-dependency list 101  # Epic issue
gh issue-dependency list 102  # Feature A
gh issue-dependency list 103  # Feature B
```

### Cross-Repository Dependencies

View dependencies that span multiple repositories:

```bash
# List dependencies for issue in shared library
gh issue-dependency list shared-lib/core#45

# Check how frontend depends on backend
gh issue-dependency list frontend-app/ui#123
```

### Integration with Other Tools

Use JSON output for integration with scripts and other tools:

```bash
# Get dependency data for processing
dependencies=$(gh issue-dependency list 123 --format json)

# Extract blocked-by count
echo "$dependencies" | jq '.blocked_by | length'

# Find all blocking issues
echo "$dependencies" | jq -r '.blocks[].number'
```

## Common Scenarios

### No Dependencies

When an issue has no dependencies:

```
Issue #123: Standalone feature implementation

No dependencies found.
```

### Mixed Repository Dependencies

When dependencies span multiple repositories:

```
Issue #123: Frontend user interface

BLOCKED BY:
  myorg/backend#45 - User API endpoint
  myorg/shared#12 - Authentication utilities

BLOCKS:
  #156 - User profile page
  myorg/mobile#78 - Mobile authentication flow
```

### Large Dependency Chains

For issues with many dependencies, the output is organized for readability:

```
Issue #123: Major refactoring epic

BLOCKED BY (3):
  #45 - Database migration
  #67 - API updates  
  #89 - Test framework setup

BLOCKS (5):
  #101 - Frontend updates
  #102 - Documentation updates
  #103 - Performance testing
  #104 - Security audit
  #105 - Release preparation
```

## Error Handling

### Issue Not Found

```
Error: Issue #999 not found in repository owner/repo

Suggestions:
- Verify the issue number exists
- Check you have access to the repository
- Ensure you're authenticated with 'gh auth login'
```

### Permission Denied

```
Error: Insufficient permissions to read issues in owner/repo

This command requires:
- Repository read access
- Issues read permission

Update your permissions or authenticate with 'gh auth login'
```

### Repository Not Found

```
Error: Repository owner/repo not found or not accessible

Suggestions:
- Verify the repository name is correct
- Check the repository is public or you have access
- Ensure you're authenticated for private repositories
```

## Related Commands

- **[`add`](add.md)** - Create new dependency relationships
- **[`remove`](remove.md)** - Remove existing dependencies

## See Also

- **[Examples](../examples/)** - Real-world usage scenarios
- **[Troubleshooting](../troubleshooting/)** - Common issues and solutions