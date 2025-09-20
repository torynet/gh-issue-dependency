# Troubleshooting

Common issues and solutions when using gh-issue-dependency.

## Installation Issues

### Command Not Found

**Problem**: `gh issue-dependency: command not found`

**Solutions**:

1. **Verify installation**:
   ```bash
   which gh-issue-dependency
   go list -m github.com/torynet/gh-issue-dependency
   ```

2. **Check PATH**:
   ```bash
   echo $PATH
   # Ensure $GOPATH/bin or $GOBIN is in your PATH
   ```

3. **Reinstall**:
   ```bash
   # For extension installation
   gh extension uninstall torynet/gh-issue-dependency
   gh extension install torynet/gh-issue-dependency
   
   # For Go installation (development)
   go clean -modcache
   go install github.com/torynet/gh-issue-dependency@latest
   ```

### GitHub CLI Not Found

**Problem**: `gh: command not found`

**Solution**: Install GitHub CLI first:
```bash
# macOS
brew install gh

# Ubuntu/Debian  
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
sudo apt update && sudo apt install gh

# Windows
winget install GitHub.cli
```

## Authentication Issues

### Not Authenticated

**Problem**: `authentication failed` or `401 Unauthorized`

**Solution**: Authenticate with GitHub CLI:
```bash
gh auth login
# Follow the interactive prompts
```

**Verify authentication**:
```bash
gh auth status
```

### Insufficient Permissions

**Problem**: `403 Forbidden` or `insufficient permissions`

**Required permissions**:
- Repository read/write access
- Issues read/write permission
- Metadata read permission

**Solutions**:

1. **Check token scopes**:
   ```bash
   gh auth token | cut -c1-10  # Show token prefix
   # Visit GitHub Settings > Developer settings > Personal access tokens
   ```

2. **Re-authenticate with correct scopes**:
   ```bash
   gh auth login --scopes repo,read:org
   ```

3. **For organization repositories**, ensure you have write access or contact an admin.

### Token Expired

**Problem**: `401 Unauthorized` after previously working

**Solution**: Refresh authentication:
```bash
gh auth refresh
```

## Command Issues

### Repository Not Found

**Problem**: `repository not found` or `404 Not Found`

**Troubleshooting steps**:

1. **Verify repository name**:
   ```bash
   gh repo view owner/repository
   ```

2. **Check current directory**:
   ```bash
   git remote -v
   # Ensure you're in the correct repository
   ```

3. **Use explicit repository flag**:
   ```bash
   gh issue-dependency list 123 --repo owner/repository
   ```

### Issue Not Found

**Problem**: `issue #123 not found`

**Troubleshooting steps**:

1. **Verify issue exists**:
   ```bash
   gh issue view 123
   ```

2. **Check issue number format**:
   ```bash
   # Correct formats:
   gh issue-dependency list 123
   gh issue-dependency list owner/repo#123
   gh issue-dependency list https://github.com/owner/repo/issues/123
   ```

3. **Verify repository context**:
   ```bash
   gh issue-dependency list 123 --repo owner/repository
   ```

### Circular Dependency Errors

**Problem**: `circular dependency detected`

**Understanding the error**:
```
❌ Cannot create dependency: circular dependency detected
   #123 → #456 → #789 → #123
```

**Solutions**:

1. **Analyze the cycle**:
   ```bash
   gh issue-dependency list 123
   gh issue-dependency list 456  
   gh issue-dependency list 789
   ```

2. **Break the cycle**:
   ```bash
   # Remove one dependency to break the cycle
   gh issue-dependency remove 789 --blocks 123
   ```

3. **Restructure dependencies**:
   - Consider if all dependencies are truly necessary
   - Look for alternative dependency structures
   - Break complex dependencies into smaller, independent pieces

## API and Network Issues

### Rate Limiting

**Problem**: `rate limit exceeded` or `403 rate limit`

**Solutions**:

1. **Wait and retry**:
   ```bash
   # GitHub API rate limits reset hourly
   sleep 60 && gh issue-dependency list 123
   ```

2. **Use authenticated requests** (higher rate limits):
   ```bash
   gh auth status  # Ensure you're authenticated
   ```

3. **Batch operations** to reduce API calls:
   ```bash
   # Instead of multiple single operations:
   gh issue-dependency add 123 --blocked-by 456,789,101
   ```

### Network Timeouts

**Problem**: `timeout` or `connection refused`

**Solutions**:

1. **Check internet connection**:
   ```bash
   curl -I https://api.github.com
   ```

2. **Retry the operation**:
   ```bash
   # The tool includes automatic retry logic
   gh issue-dependency list 123
   ```

3. **Check GitHub Status**:
   - Visit [GitHub Status](https://www.githubstatus.com/)
   - Wait for service restoration

### Corporate Firewalls

**Problem**: Connection issues in corporate environments

**Solutions**:

1. **Configure proxy** (if using corporate proxy):
   ```bash
   export HTTPS_PROXY=http://proxy.company.com:8080
   export HTTP_PROXY=http://proxy.company.com:8080
   ```

2. **Check SSL certificates**:
   ```bash
   curl -v https://api.github.com
   ```

3. **Contact IT support** for firewall configuration.

## Data and State Issues

### Inconsistent Dependency State

**Problem**: Dependencies not showing correctly or seeming inconsistent

**Troubleshooting**:

1. **Verify with GitHub web interface**:
   - Visit the issue page on GitHub
   - Check if dependencies show in the sidebar

2. **Check both sides of relationship**:
   ```bash
   gh issue-dependency list 123  # Issue that should be blocked
   gh issue-dependency list 456  # Issue that should block
   ```

3. **Refresh and retry**:
   ```bash
   # Sometimes temporary API inconsistencies resolve quickly
   sleep 10 && gh issue-dependency list 123
   ```

### Missing Dependencies

**Problem**: Expected dependencies don't appear in output

**Troubleshooting**:

1. **Check issue access**:
   ```bash
   gh issue view 123  # Can you see the issue?
   gh issue view 456  # Can you see the dependency?
   ```

2. **Verify relationship direction**:
   ```bash
   # Check both directions
   gh issue-dependency list 123  # Shows what blocks 123
   gh issue-dependency list 456  # Shows what 456 blocks
   ```

3. **Cross-repository access**:
   ```bash
   # Ensure you have access to both repositories
   gh repo view owner/repo1
   gh repo view owner/repo2
   ```

## Performance Issues

### Slow Command Execution

**Problem**: Commands taking longer than expected

**Solutions**:

1. **Check network latency**:
   ```bash
   ping api.github.com
   ```

2. **Use JSON output** for faster processing:
   ```bash
   gh issue-dependency list 123 --format json
   ```

3. **Batch operations** to reduce API calls:
   ```bash
   # More efficient:
   gh issue-dependency add 123 --blocked-by 456,789,101
   
   # Less efficient:
   gh issue-dependency add 123 --blocked-by 456
   gh issue-dependency add 123 --blocked-by 789  
   gh issue-dependency add 123 --blocked-by 101
   ```

### Memory Usage

**Problem**: High memory usage with large dependency graphs

**Solutions**:

1. **Process in smaller batches**:
   ```bash
   # Instead of processing 100 dependencies at once
   # Process in groups of 10-20
   ```

2. **Use streaming JSON processing**:
   ```bash
   gh issue-dependency list 123 --format json | jq '.blocked_by[0:10]'
   ```

## Getting Help

### Debug Information

When reporting issues, include:

1. **Version information**:
   ```bash
   gh issue-dependency --version
   gh --version
   go version
   ```

2. **Error output**:
   ```bash
   # Run with verbose output if available
   gh issue-dependency list 123 --verbose
   ```

3. **Environment information**:
   ```bash
   echo "OS: $(uname -s)"
   echo "Shell: $SHELL"  
   echo "Path: $PATH"
   ```

### Common Log Locations

Check these locations for additional error information:

- GitHub CLI logs: `~/.config/gh/`
- System logs: `/var/log/` (Linux/macOS)
- Event Viewer (Windows)

### Support Channels

1. **GitHub Issues**: [Report bugs or request features](https://github.com/torynet/gh-issue-dependency/issues)
2. **Discussions**: [Ask questions and share tips](https://github.com/torynet/gh-issue-dependency/discussions)
3. **Documentation**: [Check the latest docs](https://torynet.github.io/gh-issue-dependency/)

### Before Reporting Issues

1. **Search existing issues**: Your problem might already be reported
2. **Check the latest version**: `gh extension upgrade torynet/gh-issue-dependency` or `go install github.com/torynet/gh-issue-dependency@latest`
3. **Gather debug information**: Include version info and error messages
4. **Create minimal reproduction**: Provide steps to reproduce the issue

## Related Documentation

- **[Getting Started](../getting-started/)** - Installation and setup
- **[Commands](../commands/)** - Complete command reference
- **[Examples](../examples/)** - Usage examples and patterns