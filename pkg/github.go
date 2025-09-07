package pkg

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// RepoInfo represents repository information
type RepoInfo struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

// GitHub API client setup and repository context detection
// Following gh-sub-issue patterns for repository handling

// GetCurrentRepo gets the current repository using gh repo view --json
func GetCurrentRepo() (owner, repo string, err error) {
	// Use gh CLI to get current repository context
	cmd := exec.Command("gh", "repo", "view", "--json", "owner,name")
	output, err := cmd.Output()
	if err != nil {
		// Check if gh CLI is available and user is authenticated
		if isGhNotFound(err) {
			return "", "", NewAppError(
				ErrorTypeInternal,
				"GitHub CLI (gh) is not available",
				err,
			).WithSuggestion("Install GitHub CLI from https://cli.github.com/").
				WithSuggestion("Ensure 'gh' is in your PATH")
		}
		
		if isAuthError(err) {
			return "", "", WrapAuthError(err)
		}
		
		return "", "", NewRepositoryNotFoundError("current directory").
			WithSuggestion("Run this command from within a GitHub repository").
			WithSuggestion("Use the --repo flag to specify a repository explicitly")
	}

	// Parse the JSON response
	var repoData struct {
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
		Name string `json:"name"`
	}
	
	if err := json.Unmarshal(output, &repoData); err != nil {
		return "", "", WrapInternalError("parsing repository information", err)
	}

	return repoData.Owner.Login, repoData.Name, nil
}

// ParseRepoFlag validates and parses the --repo flag value
func ParseRepoFlag(repoFlag string) (owner, repo string, err error) {
	if repoFlag == "" {
		return "", "", NewEmptyValueError("repository")
	}

	// Handle GitHub URL format: https://github.com/owner/repo
	if strings.HasPrefix(repoFlag, "https://github.com/") {
		parts := strings.Split(strings.TrimPrefix(repoFlag, "https://github.com/"), "/")
		if len(parts) < 2 {
			return "", "", NewRepositoryFormatError(repoFlag)
		}
		return parts[0], parts[1], nil
	}

	// Handle HOST/OWNER/REPO format (rare but supported by gh CLI)
	parts := strings.Split(repoFlag, "/")
	if len(parts) == 3 {
		// Skip the host part and use owner/repo
		return parts[1], parts[2], nil
	} else if len(parts) == 2 {
		// Standard OWNER/REPO format
		return parts[0], parts[1], nil
	}

	return "", "", NewRepositoryFormatError(repoFlag).
		WithSuggestion("Use OWNER/REPO format (e.g., octocat/Hello-World)").
		WithSuggestion("Use full GitHub URL (e.g., https://github.com/octocat/Hello-World)")
}

// ParseIssueURL parses GitHub issue URLs to extract repository and issue number
func ParseIssueURL(url string) (owner, repo string, issueNumber int, err error) {
	if url == "" {
		return "", "", 0, NewEmptyValueError("issue URL")
	}

	// GitHub issue URL pattern: https://github.com/owner/repo/issues/123
	githubIssuePattern := regexp.MustCompile(`^https://github\.com/([^/]+)/([^/]+)/issues/(\d+)(?:[/?#].*)?$`)
	matches := githubIssuePattern.FindStringSubmatch(url)
	
	if matches == nil {
		return "", "", 0, NewIssueNumberValidationError(url).
			WithSuggestion("Use a GitHub issue URL (e.g., https://github.com/owner/repo/issues/123)")
	}

	owner = matches[1]
	repo = matches[2]
	issueNumber, err = strconv.Atoi(matches[3])
	if err != nil || issueNumber <= 0 {
		return "", "", 0, NewIssueNumberValidationError(url)
	}

	return owner, repo, issueNumber, nil
}

// ValidateRepoAccess validates that the user has access to the specified repository
func ValidateRepoAccess(owner, repo string) error {
	if owner == "" || repo == "" {
		return NewEmptyValueError("repository owner or name")
	}

	// Use gh CLI to check repository access
	repoName := fmt.Sprintf("%s/%s", owner, repo)
	cmd := exec.Command("gh", "repo", "view", repoName, "--json", "id")
	output, err := cmd.Output()
	if err != nil {
		if isGhNotFound(err) {
			return NewAppError(
				ErrorTypeInternal,
				"GitHub CLI (gh) is not available",
				err,
			).WithSuggestion("Install GitHub CLI from https://cli.github.com/")
		}
		
		if isAuthError(err) {
			return WrapAuthError(err)
		}
		
		// Check for specific error patterns
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "404") {
			return NewRepositoryNotFoundError(repoName)
		}
		if strings.Contains(errMsg, "forbidden") || strings.Contains(errMsg, "403") {
			return WrapPermissionError(repoName, err)
		}
		
		return NewRepositoryAccessError(repoName, err)
	}

	// Parse response to ensure we got valid repository data
	var repoData struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(output, &repoData); err != nil {
		return WrapInternalError("parsing repository validation response", err)
	}

	if repoData.ID == "" {
		return NewRepositoryAccessError(repoName, 
			fmt.Errorf("repository validation returned empty ID"))
	}

	return nil
}

// ResolveRepository resolves repository context using the priority order:
// 1. --repo flag (if provided)
// 2. Issue URL parsing (if issue is URL)  
// 3. Current repository detection via `gh repo view`
// 4. Error if no context available
func ResolveRepository(repoFlag, issueRef string) (owner, repo string, err error) {
	// Priority 1: --repo flag override
	if repoFlag != "" {
		owner, repo, err = ParseRepoFlag(repoFlag)
		if err != nil {
			return "", "", err
		}
		// Validate access to the specified repository
		if err := ValidateRepoAccess(owner, repo); err != nil {
			return "", "", err
		}
		return owner, repo, nil
	}

	// Priority 2: Issue URL parsing (if issue is URL)
	if strings.HasPrefix(issueRef, "https://github.com/") {
		var issueNumber int
		owner, repo, issueNumber, err = ParseIssueURL(issueRef)
		if err != nil {
			return "", "", err
		}
		// Validate we got a valid issue number
		if issueNumber <= 0 {
			return "", "", NewIssueNumberValidationError(issueRef)
		}
		// Validate access to the repository
		if err := ValidateRepoAccess(owner, repo); err != nil {
			return "", "", err
		}
		return owner, repo, nil
	}

	// Priority 3: Current repository detection
	owner, repo, err = GetCurrentRepo()
	if err != nil {
		return "", "", err
	}

	return owner, repo, nil
}

// Helper functions for error detection

// isGhNotFound checks if the error indicates gh CLI is not available
func isGhNotFound(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "executable file not found") ||
		   strings.Contains(errMsg, "command not found") ||
		   strings.Contains(errMsg, "not recognized as an internal")
}

// isAuthError checks if the error indicates authentication issues
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "authentication") ||
		   strings.Contains(errMsg, "unauthorized") ||
		   strings.Contains(errMsg, "to authenticate") ||
		   strings.Contains(errMsg, "gh auth login")
}

// SetupGitHubClient sets up a GitHub API client using gh CLI's authentication
func SetupGitHubClient() error {
	// Verify gh CLI is available and authenticated
	cmd := exec.Command("gh", "auth", "status")
	if err := cmd.Run(); err != nil {
		if isGhNotFound(err) {
			return NewAppError(
				ErrorTypeInternal,
				"GitHub CLI (gh) is not available",
				err,
			).WithSuggestion("Install GitHub CLI from https://cli.github.com/").
				WithSuggestion("Ensure 'gh' is in your PATH")
		}
		
		return WrapAuthError(err).
			WithSuggestion("Run 'gh auth login' to authenticate with GitHub")
	}

	return nil
}