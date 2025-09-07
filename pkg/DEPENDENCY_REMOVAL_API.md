# GitHub API Integration for Dependency Removal

This document describes the GitHub API integration implemented for deleting issue dependency relationships in the `gh-issue-dependency` extension.

## Overview

The dependency removal API provides comprehensive DELETE operations for GitHub issue dependency relationships with:
- Retry logic with exponential backoff
- Comprehensive error handling and categorization
- Batch removal operations
- Cross-repository dependency support
- User confirmation workflows
- Dry run capabilities

## Core Components

### DependencyRemover

The main struct that handles GitHub API integration for dependency removal:

```go
type DependencyRemover struct {
    client    *api.RESTClient
    validator *RemovalValidator
}
```

### Key Methods

#### `NewDependencyRemover() (*DependencyRemover, error)`
Creates a new dependency remover with GitHub API client authentication.

#### `RemoveRelationship(source, target IssueRef, relType string, opts RemoveOptions) error`
Removes a single dependency relationship between two issues with full validation pipeline.

#### `RemoveBatchRelationships(source IssueRef, targets []IssueRef, relType string, opts RemoveOptions) error`
Removes multiple dependency relationships in batch with error aggregation.

#### `RemoveCrossRepositoryRelationship(source, target IssueRef, relType string, opts RemoveOptions) error`
Handles dependency removal across different repositories with enhanced permission validation.

#### `RemoveAllRelationships(issue IssueRef, opts RemoveOptions) error`
Removes all dependency relationships (both blocked-by and blocks) for a given issue.

## API Integration Patterns

### GitHub API Endpoints

The implementation uses GitHub's dependency API endpoints:

```
DELETE /repos/{owner}/{repo}/issues/{issue_number}/dependencies/{relationship_id}
```

### Retry Logic

Implements exponential backoff with 3 retry attempts for:
- Network timeouts
- Rate limit errors (429)
- Server errors (500, 502, 503)

```go
maxRetries := 3
baseDelay := 1 * time.Second
delay := time.Duration(attempt) * baseDelay
```

### Error Handling Categories

#### Authentication Errors (401)
- Returns `WrapAuthError()` with suggestion to run `gh auth login`

#### Permission Errors (403)
- Returns `NewPermissionDeniedError()` with write permission requirements

#### Not Found Errors (404)
- Handles cases where relationships were removed by other processes

#### Rate Limiting (429)
- Automatically retries with exponential backoff

#### Network Errors
- Connection timeouts, network failures
- Automatically retries up to 3 times

## Usage Examples

### Basic Single Removal

```go
remover, err := NewDependencyRemover()
if err != nil {
    return err
}

source := CreateIssueRef("owner", "repo", 123)
target := CreateIssueRef("owner", "repo", 456)
opts := RemoveOptions{DryRun: false, Force: false}

err = remover.RemoveRelationship(source, target, "blocked-by", opts)
if err != nil {
    return err
}
```

### Batch Removal

```go
source := CreateIssueRef("owner", "repo", 123)
targets := []IssueRef{
    CreateIssueRef("owner", "repo", 456),
    CreateIssueRef("owner", "repo", 789),
}
opts := RemoveOptions{DryRun: false, Force: false}

err = remover.RemoveBatchRelationships(source, targets, "blocked-by", opts)
```

### Cross-Repository Removal

```go
source := CreateIssueRef("owner1", "repo1", 123)
target := CreateIssueRef("owner2", "repo2", 456)
opts := RemoveOptions{DryRun: false, Force: false}

err = remover.RemoveCrossRepositoryRelationship(source, target, "blocks", opts)
```

### Dry Run Mode

```go
opts := RemoveOptions{DryRun: true, Force: false}
err := remover.RemoveRelationship(source, target, "blocked-by", opts)
// Shows preview without making changes
```

## User Interface Integration

### Confirmation Prompts

When `opts.Force` is false, users see detailed confirmation prompts:

```
Remove dependency relationship?
  Source: owner/repo#123 - Feature: User Authentication System
  Target: owner/repo#456 - Database migration setup
  Type: blocked-by

This will remove the "blocked-by" relationship between these issues.
Continue? (y/N): 
```

### Success Messages

After successful removal:

```
✅ Removed blocked-by relationship: owner/repo#123 ← owner/repo#456

Dependency removed successfully.
```

### Batch Results

For batch operations:

```
✅ Removed 3 blocked-by relationships:
  owner/repo#123 ← owner/repo#456
  owner/repo#123 ← owner/repo#789
  owner/repo#123 ← owner/repo#101

Batch dependency removal completed successfully.
```

## Error Scenarios and Recovery

### Common Error Types

1. **Relationship Not Found**
   ```
   ❌ Cannot remove dependency: owner/repo#123 is not blocked by owner/repo#456
   
   Use 'gh issue-dependency list 123' to see current dependencies.
   ```

2. **Permission Denied**
   ```
   ❌ Permission denied: cannot remove dependencies in owner/repo
   
   You need write or maintain permissions to modify dependencies.
   ```

3. **Authentication Required**
   ```
   ❌ Authentication required to access GitHub
   
   Run 'gh auth login' to authenticate with GitHub.
   ```

### Batch Partial Failures

When some relationships succeed and others fail:

```
❌ Batch removal partially failed: 2 succeeded, 1 failed

Errors: owner/repo#789: Permission denied

Review the errors and retry failed operations individually.
```

## Integration Points

### Validation System (Issue #23)
- Integrates with `RemovalValidator` for comprehensive validation
- Verifies relationship existence before deletion attempts
- Checks repository permissions and issue accessibility

### Safety Features (Issue #24)
- Confirmation prompts prevent accidental deletions
- Dry run mode allows safe testing
- Force flag available for automation scenarios

### Command Interface
- Designed to integrate with Cobra CLI commands
- Supports flag-based options (--dry-run, --force)
- Compatible with existing repository context detection

## Testing

The implementation includes comprehensive tests covering:
- Struct initialization and configuration
- Error handling for various scenarios
- Utility functions and data structures
- Integration with validation components

Run tests with:
```bash
go test ./pkg -v
```

## Performance Considerations

- **Parallel API Calls**: Uses goroutines for fetching relationship data
- **Timeout Management**: 30-second timeouts for DELETE operations, 10 seconds for confirmations
- **Efficient Retry Logic**: Exponential backoff prevents API flooding
- **Batch Operations**: Single validation pass for multiple targets

## Security Features

- **Authentication Verification**: Validates GitHub CLI auth before operations
- **Permission Checking**: Verifies write access to repositories
- **Cross-Repository Validation**: Enhanced permission checks for cross-repo operations
- **Safe Defaults**: Confirmation required unless explicitly forced