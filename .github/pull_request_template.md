## Summary
Brief description of the changes in this PR.

**Issue:** #{issue_number}

## Type of Change
- [ ] 🐛 Bug fix (patch version bump)
- [ ] ✨ New feature (minor version bump) 
- [ ] ⚡ Performance improvement (patch version bump)
- [ ] 🔄 Refactoring (no version bump)
- [ ] 📚 Documentation update (no version bump)
- [ ] 🧪 Test improvements (no version bump)
- [ ] 🔧 Build/CI changes (no version bump)

## Breaking Changes
- [ ] ⚠️ **This PR contains breaking changes** (major version bump)

**If breaking changes are checked, complete the following:**

### Breaking Change Description:
<!-- Describe what existing functionality will break and why this change is necessary -->


### Migration Guide:  
<!-- Provide clear steps for users to migrate from the old behavior to the new behavior -->
1. 
2. 
3. 

### Affected APIs/Functions:
<!-- List the specific APIs, functions, or interfaces that are changing -->
- 
- 

---

## Testing
- [ ] Unit tests added/updated and passing
- [ ] Integration tests added/updated and passing  
- [ ] Manual testing completed
- [ ] Cross-platform compatibility verified (if applicable)

## Conventional Commit Preview
Based on your selections above, the squash merge will create a conventional commit like:

**For regular changes:**
- Bug fix: `fix: {PR title}`
- New feature: `feat: {PR title}`
- Performance: `perf: {PR title}`

**For breaking changes:**
- Breaking feature: `feat!: {PR title}`
- Breaking fix: `fix!: {PR title}`

The commit message will also include:
```
BREAKING CHANGE: {Breaking Change Description from above}
```

## PR Title Format

**New format (recommended for automatic releases):**
```
{issue#}: {type}: {description}
```

**Examples:**
- `27: fix: improve error handling in API responses`
- `34: feat: add user authentication with OAuth`
- `42: perf: optimize database queries for large datasets`
- `18: feat!: breaking change to configuration API`

**Legacy format (deprecated but still accepted):**
```
{issue#}: {description}
```

**Valid conventional commit types:**
- `feat`: New features (minor version bump)
- `fix`: Bug fixes (patch version bump)  
- `perf`: Performance improvements (patch version bump)
- `docs`: Documentation changes (no version bump)
- `style`: Code style changes (no version bump)
- `refactor`: Code refactoring (no version bump)
- `test`: Test changes (no version bump)
- `chore`: Maintenance tasks (no version bump)
- `ci`: CI/CD changes (no version bump)
- `build`: Build system changes (no version bump)
- `revert`: Revert previous changes (no version bump)

Add `!` after the type for breaking changes: `feat!:` or `fix!:`

## Pre-merge Checklist
- [ ] Branch name follows convention: `(feature|hotfix|epic)/{issue#}-description`
- [ ] PR title follows new format: `{issue#}: {type}: {description}` (or legacy format during transition)
- [ ] All CI checks are passing (branch validation, tests, linting)
- [ ] Code follows project conventions and style guide
- [ ] Self-review completed
- [ ] Breaking changes properly documented (if applicable)
- [ ] Ready for squash merge to main

---

## Additional Notes
<!-- Any additional context, screenshots, or information that reviewers should know -->

## Reviewer Guidelines
When approving this PR for merge, please ensure:
1. The squash merge commit message follows conventional commit format
2. Breaking changes are properly indicated with `!` or `BREAKING CHANGE:`
3. The commit type (`feat`, `fix`, `perf`) matches the change type selected above