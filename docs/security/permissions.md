# Security & Permissions

## Overview

The `gh-issue-dependency` extension requires minimal GitHub permissions to function securely while providing complete dependency management capabilities.

## Authentication Method

### GitHub CLI Integration
- **Authentication**: Delegates to GitHub CLI (`gh auth`)
- **Token Storage**: Uses GitHub CLI's secure token storage (OS keychain)
- **No Direct Token Handling**: Extension never stores or manages tokens directly
- **Session Management**: Inherits GitHub CLI's authentication state

### Security Benefits
- ✅ **Zero Token Exposure**: Extension never handles raw tokens
- ✅ **Secure Storage**: Leverages OS-level credential management
- ✅ **Standard Authentication**: Uses GitHub's official CLI authentication
- ✅ **Session Isolation**: Each user session is independently authenticated

## Required GitHub Permissions

### Minimum Repository Permissions

| Operation | Required Permission | Scope | Justification |
|-----------|-------------------|-------|---------------|
| **List Dependencies** | `READ` | Repository metadata, Issues | View existing issue dependency relationships |
| **Add Dependencies** | `WRITE` | Issues | Create new dependency relationships between issues |
| **Remove Dependencies** | `WRITE` | Issues | Delete existing dependency relationships |
| **Cross-Repository** | `READ` (target repo) | Repository metadata | Validate target repository exists and is accessible |

### Permission Verification
The extension automatically validates permissions before attempting operations:
- Repository access validation via `repos/{owner}/{repo}` endpoint
- Permission level checking through repository API response
- Graceful degradation with helpful error messages for insufficient permissions

## API Endpoints Used

### Read Operations (Require READ permission)
```
GET /repos/{owner}/{repo}
GET /repos/{owner}/{repo}/issues/{issue_number}
GET /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocked_by
GET /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocks
```

### Write Operations (Require WRITE permission)
```
DELETE /repos/{owner}/{repo}/issues/{issue_number}/dependencies/{relationship_id}
```

## Security Controls

### Input Validation
- ✅ Repository name format validation (`owner/repo`)
- ✅ Issue number validation (positive integers only)
- ✅ URL parsing with strict GitHub.com domain requirements
- ✅ Relationship type validation (blocked_by, blocks only)

### Error Handling
- ✅ **No Credential Leakage**: Error messages never expose authentication tokens
- ✅ **Safe Error Patterns**: Generic error messages for security-sensitive failures
- ✅ **Structured Error Types**: Categorized errors (auth, permission, network, validation)
- ✅ **User Guidance**: Helpful suggestions without exposing implementation details

### Network Security
- ✅ **HTTPS Only**: All API communication over encrypted connections
- ✅ **Official Endpoints**: Only communicates with api.github.com
- ✅ **Request Validation**: All API requests validated before transmission
- ✅ **Timeout Controls**: 30-second timeouts prevent hanging connections

### Rate Limiting
- ✅ **Automatic Retry**: Exponential backoff for rate-limited requests
- ✅ **Respectful Usage**: Follows GitHub's rate limiting guidelines
- ✅ **Error Recovery**: Graceful handling of rate limit responses

## Permission Models

### Organization Repositories
For organization-owned repositories, users need:
- **Member**: Minimum `triage` role for read operations
- **Collaborator**: Minimum `write` role for dependency modifications
- **Admin**: Full access to all dependency operations

### Personal Repositories
For user-owned repositories:
- **Owner**: Full access to all operations
- **Collaborator**: `write` permission required for modifications
- **Read-Only**: Can list dependencies but cannot modify

### Cross-Repository Dependencies
When creating dependencies across repositories:
- **Source Repository**: Requires `write` permission (where dependency is created)
- **Target Repository**: Requires `read` permission (repository being referenced)
- **Validation**: Both repositories must be accessible to authenticated user

## Security Best Practices

### For Users
1. **Use Organization Tokens**: For organization repositories, use tokens with appropriate scope
2. **Regular Auth Review**: Periodically review GitHub CLI authentication with `gh auth status`
3. **Minimal Scope**: Grant only necessary repository access
4. **Token Rotation**: Follow your organization's token rotation policies

### For Administrators
1. **Repository Policies**: Set appropriate default permissions for repositories
2. **Audit Trails**: Monitor dependency modifications through GitHub's audit logs
3. **Branch Protection**: Consider protecting important branches from unauthorized changes
4. **Access Reviews**: Regularly review repository collaborator permissions

## Compliance Features

### Data Privacy
- ✅ **No Data Storage**: Extension stores no user data or repository information
- ✅ **Ephemeral Operations**: All operations are stateless and immediate
- ✅ **Local Cache Only**: Optional local caching for performance (can be disabled)

### Audit Support
- ✅ **GitHub Audit Logs**: All API operations appear in GitHub's audit logs
- ✅ **Structured Logging**: Extension operations can be logged for compliance
- ✅ **Change Attribution**: All changes attributed to authenticated user
- ✅ **Operation Tracking**: Clear audit trail for all dependency modifications

## Common Security Scenarios

### Authentication Failures
```bash
# Error: Authentication required
$ gh issue-dependency list 123
❌ Authentication required to access GitHub

Run 'gh auth login' to authenticate with GitHub.
```

### Permission Denied
```bash
# Error: Insufficient permissions
$ gh issue-dependency remove 123 --blocked-by 456
❌ Insufficient permissions to access owner/repo

Ensure you have at least write permissions for this repository.
You may need to be added as a collaborator or team member.
```

### Cross-Repository Access
```bash
# Error: Cannot access target repository
$ gh issue-dependency add 123 --blocks other-org/private-repo#456
❌ Cannot access target repository other-org/private-repo

Verify you have access to this repository.
Check if the repository is private and you're authenticated.
```

## Security Incident Response

### Token Compromise
If a GitHub token is compromised:
1. **Revoke Token**: Use `gh auth logout` to revoke current session
2. **Re-authenticate**: Run `gh auth login` with new credentials
3. **Review Activity**: Check GitHub's audit logs for unauthorized activity
4. **Update Access**: Review and update repository permissions as needed

### Unauthorized Changes
If unauthorized dependency changes are detected:
1. **Audit Logs**: Review GitHub repository audit logs
2. **Revert Changes**: Use GitHub's issue interface to restore dependency relationships
3. **Permission Review**: Audit repository collaborator permissions
4. **Process Update**: Update access control procedures if needed

## Additional Resources

- [GitHub CLI Authentication](https://cli.github.com/manual/gh_auth)
- [GitHub Repository Permissions](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/managing-repository-settings/managing-teams-and-people-with-access-to-your-repository)
- [GitHub API Security](https://docs.github.com/en/rest/overview/other-authentication-methods)
- [Organization Security Best Practices](https://docs.github.com/en/organizations/keeping-your-organization-secure)