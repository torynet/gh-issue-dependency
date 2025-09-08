# remove command

Remove dependency relationships between GitHub issues.

## Synopsis

```bash
gh issue-dependency remove <issue> --blocked-by <target-issue> [flags]
gh issue-dependency remove <issue> --blocks <target-issue> [flags]
gh issue-dependency remove <issue> --all [flags]
```

## Description

The `remove` command safely removes dependency relationships between issues. It includes safety features like confirmation prompts and dry-run mode to prevent accidental deletions of important dependency relationships.

## Usage

### Basic Usage

```bash
# Remove a specific blocked-by relationship
gh issue-dependency remove 123 --blocked-by 456

# Remove a specific blocks relationship  
gh issue-dependency remove 123 --blocks 789

# Remove all dependencies for an issue
gh issue-dependency remove 123 --all
```

### Multiple Dependencies

```bash
# Remove multiple blocked-by relationships at once
gh issue-dependency remove 123 --blocked-by 456,789,101

# Remove multiple blocks relationships
gh issue-dependency remove 123 --blocks 456,789,101
```

### Cross-Repository Dependencies

```bash
# Remove cross-repository dependency
gh issue-dependency remove 123 --blocked-by owner/other-repo#456

# Use with explicit repository specification
gh issue-dependency remove 123 --blocked-by 456 --repo myorg/myproject
```

## Flags

### `--blocked-by <issue-list>`
Remove specific blocked-by relationships. The target issue will no longer be blocked by these issues.

### `--blocks <issue-list>`
Remove specific blocks relationships. These issues will no longer be blocked by the target issue.

### `--all`
Remove all dependency relationships for the issue (both blocked-by and blocks).

### `--dry-run`
Preview what would be removed without actually making changes.

### `--force`
Skip confirmation prompts. Use with caution, especially in automation.

### `--repo <owner/repo>`
Repository to use when not in a git repository.

### `--help`
Show help for the remove command.

## Safety Features

### Interactive Confirmation

By default, the command asks for confirmation before removing dependencies:

```bash
gh issue-dependency remove 123 --blocked-by 456
```

```
⚠️  Remove dependency relationship?

  Source: myorg/myproject #123 - Implement user authentication system
  Target: myorg/myproject #456 - Set up database schema
  Type: blocked-by

This will remove the "blocked-by" relationship between these issues.
Continue? (y/N): 
```

### Dry-Run Mode

Preview changes before making them:

```bash
gh issue-dependency remove 123 --blocked-by 456 --dry-run
```

```
🔍 Dry run: dependency removal preview

Would remove:
  ❌ blocked-by relationship: #123 ← #456 (Set up database schema)
     This means #123 would no longer be blocked by #456

💡 No changes made. Use --force to skip confirmation or remove --dry-run to execute.
```

### Batch Operation Warnings

When removing multiple dependencies, additional warnings are shown:

```bash
gh issue-dependency remove 123 --all --dry-run
```

```
⚠️ WARNING: Batch removal of 5 relationships
   This action will remove multiple dependency relationships at once.
   Consider using --dry-run first to preview all changes.

Would remove:
  ❌ blocked-by relationship: #123 ← #456 (Database setup)
  ❌ blocked-by relationship: #123 ← #789 (User model)  
  ❌ blocks relationship: #123 → #101 (Login form)
  ❌ blocks relationship: #123 → #102 (Password reset)
  ❌ blocks relationship: #123 → #103 (User profile)
```

## Examples

### Removing Obsolete Dependencies

```bash
# Remove dependency that's no longer needed
gh issue-dependency remove 123 --blocked-by 456

# Remove multiple obsolete dependencies
gh issue-dependency remove 123 --blocked-by 456,789 --dry-run  # Preview first
gh issue-dependency remove 123 --blocked-by 456,789           # Then execute
```

### Restructuring Dependencies

```bash
# Remove old dependency structure
gh issue-dependency remove 100 --all --dry-run  # Preview removal
gh issue-dependency remove 100 --all            # Remove all

# Add new dependency structure  
gh issue-dependency add 100 --blocked-by 200,201,202
```

### Cross-Repository Cleanup

```bash
# Remove cross-repository dependency that's no longer valid
gh issue-dependency remove frontend/ui#123 --blocked-by backend/api#456

# Clean up dependencies in shared repository
gh issue-dependency remove shared/core#100 --blocks frontend/ui#123,mobile/ios#789
```

### Sprint Completion Cleanup

```bash
# Remove completed epic dependencies
gh issue-dependency remove 100 --blocked-by 101,102,103  # Features are done

# Remove sprint milestone dependencies
gh issue-dependency remove 200 --all  # Sprint is complete
```

## Output Examples

### Successful Removal

```bash
gh issue-dependency remove 123 --blocked-by 456
```

```
✅ Removed blocked-by relationship: #123 ← #456 (Database migration setup)

Dependency removed successfully.
```

### Multiple Dependencies Removed

```bash
gh issue-dependency remove 123 --blocked-by 456,789,101
```

```
✅ Removed blocked-by relationships:
   #123 ← #456 (Database migration setup)  
   #123 ← #789 (User model creation)
   #123 ← #101 (Authentication framework)

Dependencies removed successfully (3 relationships).
```

### All Dependencies Removed

```bash
gh issue-dependency remove 123 --all
```

```
✅ Removed all dependency relationships for issue #123:
   Blocked-by relationships (2):
     #123 ← #456 (Database setup)
     #123 ← #789 (User model)
   
   Blocks relationships (3):  
     #123 → #101 (Login form)
     #123 → #102 (Password reset)
     #123 → #103 (User profile)

All dependencies removed successfully (5 relationships).
```

### User Cancellation

```bash
gh issue-dependency remove 123 --blocked-by 456
# User responds "n" to confirmation prompt
```

```
❌ Operation cancelled by user.

No changes were made to issue dependencies.
```

## Error Handling

### Relationship Not Found

```
❌ Cannot remove dependency: relationship does not exist

Issue #123 is not blocked by #456.

Use 'gh issue-dependency list 123' to see existing dependencies.
```

### Permission Errors

```
❌ Cannot remove dependency: insufficient permissions

Required permissions:
- Write access to repository (myorg/myproject)
- Issues: write permission

Contact a repository administrator to update your permissions.
```

### Issue Not Found

```
❌ Cannot remove dependency: issue not found

Issue #999 does not exist in repository myorg/myproject.

Suggestions:
- Verify the issue number is correct  
- Check the issue hasn't been deleted
- Ensure you have access to view the issue
```

## Force Mode

For automation scenarios, use `--force` to skip confirmations:

```bash
# Skip confirmation (use carefully)
gh issue-dependency remove 123 --blocked-by 456 --force

# Useful in scripts and CI/CD
gh issue-dependency remove 123 --all --force
```

⚠️ **Warning**: Force mode skips all safety confirmations. Use only when you're certain about the changes, especially in automation scenarios.

## Best Practices

### Before Removing Dependencies

1. **Understand impact**: Use `gh issue-dependency list` to see current relationships
2. **Preview changes**: Always use `--dry-run` first for complex removals
3. **Team communication**: Inform team members about dependency structure changes
4. **Document reasons**: Add comments to issues explaining why dependencies were removed

### Safe Removal Workflow

1. **Analyze current state**:
   ```bash
   gh issue-dependency list 123
   ```

2. **Preview removal**:
   ```bash
   gh issue-dependency remove 123 --blocked-by 456 --dry-run
   ```

3. **Execute with confirmation**:
   ```bash
   gh issue-dependency remove 123 --blocked-by 456
   ```

4. **Verify result**:
   ```bash
   gh issue-dependency list 123
   ```

### Managing Complex Removals

1. **One relationship at a time**: For critical dependencies, remove one at a time
2. **Batch similar removals**: Group related dependency removals together
3. **Coordinate timing**: Remove dependencies when it won't block active work
4. **Backup strategy**: Document removed dependencies in case they need to be restored

### Automation Guidelines

1. **Use dry-run first**: Always test automated removal scripts with `--dry-run`
2. **Error handling**: Check exit codes and handle errors gracefully
3. **Logging**: Log all dependency changes for audit trails
4. **Rollback plan**: Have a process to restore dependencies if needed

## Related Commands

- **[`list`](list.md)** - View existing dependencies before removal
- **[`add`](add.md)** - Create new dependency relationships

## See Also

- **[Examples](../examples/)** - Real-world usage scenarios  
- **[Troubleshooting](../troubleshooting/)** - Common issues and solutions