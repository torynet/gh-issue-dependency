# add command

Create dependency relationships between GitHub issues.

## Synopsis

```bash
gh issue-dependency add <issue> --blocked-by <target-issue> [flags]
gh issue-dependency add <issue> --blocks <target-issue> [flags]
```

## Description

The `add` command creates dependency relationships between issues. You can specify that an issue is "blocked by" other issues (dependencies that must be completed first) or that an issue "blocks" other issues (issues that cannot start until this one is complete).

## Usage

### Basic Usage

```bash
# Make issue #123 blocked by issue #456
gh issue-dependency add 123 --blocked-by 456

# Make issue #123 block issue #789  
gh issue-dependency add 123 --blocks 789

# Use GitHub URLs
gh issue-dependency add https://github.com/owner/repo/issues/123 --blocked-by 456
```

### Multiple Dependencies

```bash
# Add multiple blocked-by relationships at once
gh issue-dependency add 123 --blocked-by 456,789,101

# Add multiple blocks relationships
gh issue-dependency add 123 --blocks 456,789,101
```

### Cross-Repository Dependencies

```bash
# Issue #123 is blocked by issue #456 in another repository
gh issue-dependency add 123 --blocked-by owner/other-repo#456

# Use with explicit repository specification
gh issue-dependency add 123 --blocked-by 456 --repo myorg/myproject
```

## Flags

### `--blocked-by <issue-list>`
Specify issues that block the target issue. The target issue cannot be completed until these issues are resolved.

### `--blocks <issue-list>`  
Specify issues that are blocked by the target issue. These issues cannot start until the target issue is completed.

### `--dry-run`
Preview the changes without actually creating the dependencies.

### `--repo <owner/repo>`
Repository to use when not in a git repository.

### `--help`
Show help for the add command.

## Relationship Types

### blocked-by Relationships

When you specify `--blocked-by`, you're creating a dependency where the target issue cannot proceed until the specified issues are complete.

```bash
# Issue #123 cannot start until #456 and #789 are done
gh issue-dependency add 123 --blocked-by 456,789
```

### blocks Relationships  

When you specify `--blocks`, you're creating a dependency where the specified issues cannot proceed until the target issue is complete.

```bash
# Issues #456 and #789 cannot start until #123 is done
gh issue-dependency add 123 --blocks 456,789
```

## Validation and Safety

### Circular Dependency Prevention

The command automatically detects and prevents circular dependencies:

```bash
# This would fail if it creates a circular dependency
gh issue-dependency add 123 --blocked-by 456
# Error: Cannot create dependency: circular dependency detected
# #123 → #456 → #789 → #123
```

### Dry-Run Mode

Preview changes before making them:

```bash
gh issue-dependency add 123 --blocked-by 456 --dry-run
```

```
Dry run: dependency creation preview

Would create:
  ✓ blocked-by relationship: #123 ← #456 (Database setup task)

Validation checks:
  ✓ Issues exist and are accessible  
  ✓ User has write permissions
  ✓ No circular dependency detected
  ✓ Relationship does not already exist

Use --force to skip confirmation or remove --dry-run to execute.
```

### Duplicate Detection

The command detects existing relationships:

```
Error: Cannot create dependency: relationship already exists
#123 is already blocked by #456

Use 'gh issue-dependency list 123' to see existing dependencies.
```

## Examples

### Sprint Planning Workflow

```bash
# Set up epic dependencies
gh issue-dependency add 100 --blocked-by 101,102,103  # Epic waits for features

# Create feature dependency chain  
gh issue-dependency add 101 --blocked-by 104          # Feature waits for API
gh issue-dependency add 102 --blocked-by 104          # Feature waits for API  
gh issue-dependency add 103 --blocked-by 101,102      # Feature waits for other features
```

### Release Preparation

```bash
# Release issue blocked by all required features
gh issue-dependency add 200 --blocked-by 201,202,203,204

# Testing blocked by feature completion
gh issue-dependency add 300 --blocked-by 200

# Documentation updates can happen in parallel
gh issue-dependency add 400 --blocked-by 201,202,203,204
```

### Cross-Team Coordination

```bash
# Frontend blocked by backend API
gh issue-dependency add frontend/ui#123 --blocked-by backend/api#456

# Mobile app blocked by shared library updates  
gh issue-dependency add mobile/ios#789 --blocked-by shared/core#101
```

### Database Migration Dependencies

```bash
# Application changes blocked by database migration
gh issue-dependency add 301 --blocked-by 300  # App change waits for migration
gh issue-dependency add 302 --blocked-by 300  # Another change waits for migration
gh issue-dependency add 303 --blocked-by 300  # Third change waits for migration

# Migration blocks deployment
gh issue-dependency add 300 --blocks 400      # Migration blocks deployment
```

## Output Examples

### Successful Creation

```bash
gh issue-dependency add 123 --blocked-by 456
```

```
✅ Added blocked-by relationship: #123 ← #456 (Database migration setup)

Dependency created successfully.
```

### Multiple Dependencies

```bash  
gh issue-dependency add 123 --blocked-by 456,789,101
```

```
✅ Added blocked-by relationships:
   #123 ← #456 (Database migration setup)
   #123 ← #789 (User model creation)  
   #123 ← #101 (Authentication framework)

Dependencies created successfully (3 relationships).
```

### Cross-Repository Creation

```bash
gh issue-dependency add 123 --blocked-by myorg/backend#456
```

```
✅ Added blocked-by relationship: myorg/frontend#123 ← myorg/backend#456

Cross-repository dependency created successfully.
```

## Error Handling

### Permission Errors

```
❌ Cannot create dependency: insufficient permissions

Required permissions:
- Write access to source repository (myorg/frontend)  
- Write access to target repository (myorg/backend)
- Issues: write permission

Update your permissions or contact a repository administrator.
```

### Issue Not Found

```
❌ Cannot create dependency: issue not found

Issue #999 does not exist in repository myorg/myproject.

Suggestions:
- Verify the issue number is correct
- Check the issue hasn't been deleted
- Ensure you have access to view the issue
```

### Validation Errors

```
❌ Cannot create dependency: circular dependency detected
   #123 → #456 → #789 → #123

This would create a circular dependency. Consider:
- Removing the dependency from #789 to #123, or  
- Restructuring the dependency relationships
```

## Best Practices

### Planning Dependencies

1. **Map workflow first**: Understand your team's process before creating dependencies
2. **Avoid over-constraining**: Don't create unnecessary dependencies that slow progress
3. **Use epics wisely**: Create logical groupings with clear dependency chains
4. **Consider parallelization**: Look for work that can happen simultaneously

### Managing Complex Dependencies

1. **Use dry-run mode**: Always preview complex dependency changes
2. **Document rationale**: Use issue comments to explain dependency decisions
3. **Regular review**: Periodically review and clean up outdated dependencies
4. **Team communication**: Ensure all team members understand the dependency structure

### Cross-Repository Dependencies

1. **Minimize coupling**: Reduce cross-repository dependencies when possible
2. **Clear communication**: Ensure teams understand cross-repository impacts
3. **Version coordination**: Consider how dependencies affect release coordination
4. **Access management**: Ensure teams have appropriate repository access

## Related Commands

- **[`list`](list.md)** - View existing dependencies
- **[`remove`](remove.md)** - Remove dependency relationships

## See Also

- **[Examples](../examples/)** - Real-world usage scenarios
- **[Troubleshooting](../troubleshooting/)** - Common issues and solutions