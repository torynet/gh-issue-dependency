# Documentation

This directory contains documentation for the gh-issue-dependency project's development and release processes.

## Documentation Index

### Development
- **[Development Workflow](development-workflow.md)** - Complete guide for developers
  - Branch naming conventions
  - Commit message guidelines  
  - PR process and requirements
  - Local development setup
  - Testing requirements
  - Best practices and troubleshooting

### Release Management  
- **[Release Pipeline](release-pipeline.md)** - Complete release automation system
  - Trunk-based development workflow
  - Conventional commits and versioning
  - GitHub Actions workflows
  - Manual approval gates
  - Build artifacts and deployment
  - Environment configuration

## Quick Reference

### New Developer Onboarding
1. Read [Development Workflow](development-workflow.md)
2. Run `./scripts/install-hooks.sh` to set up git hooks
3. Create first feature branch: `feature/{issue#}-{description}`
4. Follow PR template for submitting changes

### Release Process Overview
1. **Development**: Work in feature branches
2. **Integration**: Squash merge to main with conventional commits
3. **RC Creation**: Automatic RC tag and beta build
4. **Testing**: Manual validation of RC artifacts  
5. **Release**: Manual approval creates production release

### Branch Naming
```
feature/{issue-number}-{description}
hotfix/{issue-number}-{description}  
epic/{issue-number}-{description}
```

### Conventional Commits
- `feat:` → Minor version bump (new features)
- `fix:` → Patch version bump (bug fixes)
- `feat!:` → Major version bump (breaking changes)
- `docs:`, `style:`, `refactor:`, `test:`, `build:`, `ci:`, `chore:` → No version bump

## Pipeline Architecture

```mermaid
graph LR
    A[Feature Branch] --> B[Pull Request]
    B --> C[Squash Merge to Main]
    C --> D[RC Tag Created]
    D --> E[Beta Build]
    E --> F[Manual Approval Gate]
    F --> G[Release Tag]
    G --> H[Production Build]
    H --> I[GitHub Release]
```

## Support

For questions about the development workflow or release process:
1. Check the relevant documentation above
2. Review GitHub Actions workflow logs
3. Check existing GitHub issues
4. Create new issue for workflow problems