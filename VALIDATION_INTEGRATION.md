# Dependency Remove Validation Integration

This document explains how the validation system implemented in Issue #23 integrates with the remove command structure from Issue #22.

## Overview

The validation system provides comprehensive validation for dependency removal operations, including:

- **Relationship existence verification** using GitHub API
- **Permission checking** for repository write access  
- **Issue accessibility validation** for both source and target issues
- **Integration with existing validation patterns** from dependency-add epic
- **Clear error messages** when relationships don't exist
- **Batch validation** for multiple target issues

## Core Components

### RemovalValidator

The main validation class that orchestrates all validation operations:

```go
type RemovalValidator struct {
    client *api.RESTClient
}

func NewRemovalValidator() (*RemovalValidator, error)
func (v *RemovalValidator) ValidateRemoval(source, target IssueRef, relType string) error
func (v *RemovalValidator) ValidateBatchRemoval(source IssueRef, targets []IssueRef, relType string) error  
func (v *RemovalValidator) VerifyRelationshipExists(source, target IssueRef, relType string) (bool, error)
```

### IssueRef

Structured representation of issue references supporting cross-repository operations:

```go
type IssueRef struct {
    Owner    string
    Repo     string  
    Number   int
    FullName string // owner/repo format
}
```

### RemoveOptions

Options controlling removal behavior:

```go
type RemoveOptions struct {
    DryRun bool
    Force  bool
}
```

## Integration with cmd/remove.go

The validation system integrates seamlessly with the existing remove command structure:

```go
// In cmd/remove.go RunE function:
func(cmd *cobra.Command, args []string) error {
    // 1. Create validation executor
    executor, err := pkg.NewRemovalExecutor()
    if err != nil {
        return err
    }

    // 2. Set up options
    opts := pkg.RemoveOptions{
        DryRun: dryRun,
        Force:  force,
    }

    // 3. Execute with validation
    if removeAll {
        return executor.ExecuteRemovalAll(args[0], opts)
    } else if removeBlockedBy != "" {
        return executor.ExecuteRemoval(args[0], removeBlockedBy, "blocked-by", opts)
    } else if removeBlocks != "" {
        return executor.ExecuteRemoval(args[0], removeBlocks, "blocks", opts)
    }

    return pkg.NewAppError(pkg.ErrorTypeValidation, "No removal type specified", nil)
}
```

## Validation Flow

### Single Dependency Removal

1. **Input Validation**: Parse and validate issue references
2. **Permission Checking**: Verify write access to source repository
3. **Issue Access Validation**: Confirm both issues exist and are accessible
4. **Relationship Verification**: Check that the dependency relationship exists
5. **User Confirmation**: Prompt unless `--force` flag is used
6. **Execution**: Perform actual removal via GitHub API

### Batch Dependency Removal

For multiple targets (comma-separated list):

1. **Source Validation**: Validate source issue and permissions once
2. **Target Validation**: Validate each target issue individually  
3. **Relationship Verification**: Check each relationship exists
4. **Aggregate Results**: Collect all validation errors
5. **Batch Confirmation**: Single confirmation for all changes
6. **Batch Execution**: Remove all valid relationships

## Error Scenarios and Messages

### Non-Existent Relationship

```text
❌ Cannot remove dependency: #123 is not blocked by #456

Details:
  source: owner/repo#123
  target: owner/repo#456
  relationship_type: blocked-by

Suggestions:
  • Use 'gh issue-dependency list 123' to see current dependencies
  • Verify the issue numbers and relationship type are correct
```

### Permission Denied

```text
❌ Permission denied: cannot modify dependencies in owner/repo

Details:
  operation: modify dependencies
  repository: owner/repo

Suggestions:
  • Ensure you have write or maintain permissions for this repository
  • Contact the repository owner to request appropriate access
```

### Issue Not Found

```text
❌ Issue #123 not found in owner/repo

Details:
  repository: owner/repo
  issue_number: 123

Suggestions:
  • Verify the issue number exists in the repository
  • Check if you have access to view the issue
```

## API Efficiency

The validation system minimizes API calls through:

- **Parallel API requests** for fetching issue details and dependencies
- **Caching of dependency data** using existing patterns from github.go
- **Batch validation** reducing redundant permission checks
- **Fallback strategies** for permission validation

## Testing

Comprehensive test coverage includes:

- **Unit tests** for validation logic (`validation_test.go`)
- **Integration examples** showing command integration
- **Error scenario validation** for all error types
- **Mock-friendly design** for testing without API calls

## Usage Examples

### Basic Removal with Validation

```bash
# Remove a single dependency
gh issue-dependency remove 123 --blocked-by 456

# Validation checks:
# 1. Issue #123 exists and is accessible
# 2. Issue #456 exists and is accessible  
# 3. User has write permissions to repository
# 4. blocked-by relationship exists between #123 and #456
# 5. User confirms removal (unless --force)
```

### Batch Removal with Validation

```bash
# Remove multiple dependencies  
gh issue-dependency remove 123 --blocks 456,789,101

# Validation checks:
# 1. Issue #123 exists and has write permissions
# 2. Issues #456, #789, #101 all exist and are accessible
# 3. All three "blocks" relationships exist
# 4. User confirms batch removal (unless --force)
```

### Dry Run Mode

```bash
# Preview what would be removed
gh issue-dependency remove 123 --blocked-by 456 --dry-run

# Output:
# Dry run: dependency removal preview
#
# Would remove:
#   ❌ blocked-by relationship: owner/repo#123 ← owner/repo#456
#
# No changes made. Use --force to skip confirmation or remove --dry-run to execute.
```

## Error Recovery Guidance

All validation errors provide specific suggestions for resolution:

- **Permission errors**: Direct to repository settings or owner contact
- **Issue not found**: Suggest verification and access checks
- **Relationship not found**: Show current dependencies with list command
- **API errors**: Provide GitHub status and retry guidance

## Integration Benefits

1. **Safety First**: Prevents accidental removal of non-existent relationships
2. **Clear Feedback**: Specific error messages guide user to resolution
3. **Efficiency**: Minimal API calls through intelligent batching and caching
4. **Consistency**: Reuses established patterns from dependency-add and list commands
5. **Extensibility**: Easy to add new validation rules and error scenarios

This validation system ensures reliable, user-friendly dependency removal with comprehensive error handling and clear guidance for all scenarios.