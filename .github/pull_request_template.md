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

## Pre-merge Checklist
- [ ] Branch name follows convention: `(feature|hotfix|epic)/{issue#}-description`
- [ ] PR title includes issue number: `{issue#}: Description`
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