// Package pkg provides shared utilities and types for the gh-issue-dependency extension.
//
// This package contains GitHub API integration, error handling, repository context
// detection, and other common functionality used across all commands.
package pkg

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
)

// RepoInfo represents repository information returned from GitHub API calls.
// This structure is used for JSON unmarshaling of repository data.
type RepoInfo struct {
	Owner string `json:"owner"` // Repository owner (user or organization)
	Name  string `json:"name"`  // Repository name
}

// GitHub API Data Structures for Issue Dependencies
//
// These structures model the GitHub API responses for issue dependency relationships
// and issue details. They are used for marshaling API responses and providing
// structured data to the output formatting system.

// User represents a GitHub user or organization
type User struct {
	Login   string `json:"login"`
	HTMLURL string `json:"html_url"`
}

// Label represents a GitHub issue label
type Label struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// Issue represents a GitHub issue with dependency-relevant fields
type Issue struct {
	Number    int      `json:"number"`
	Title     string   `json:"title"`
	State     string   `json:"state"`
	Assignees []User   `json:"assignees"`
	Labels    []Label  `json:"labels"`
	HTMLURL   string   `json:"html_url"`
	Repository string  `json:"repository,omitempty"` // Added for cross-repo dependencies
}

// DependencyRelation represents a relationship between issues
type DependencyRelation struct {
	Issue      Issue  `json:"issue"`
	Type       string `json:"type"`       // "blocked_by" or "blocks"
	Repository string `json:"repository"` // Repository of the related issue
}

// DependencyData contains all dependency information for an issue
type DependencyData struct {
	SourceIssue             Issue                 `json:"source_issue"`
	BlockedBy               []DependencyRelation  `json:"blocked_by"`
	Blocking                []DependencyRelation  `json:"blocking"`
	FetchedAt               time.Time            `json:"fetched_at"`
	TotalCount              int                  `json:"total_count"`
	OriginalBlockedByCount  int                  `json:"original_blocked_by_count,omitempty"`
	OriginalBlockingCount   int                  `json:"original_blocking_count,omitempty"`
}

// CacheEntry represents a cached dependency data entry
type CacheEntry struct {
	Data      DependencyData `json:"data"`
	ExpiresAt time.Time      `json:"expires_at"`
}

// Cache configuration
const (
	CacheDir     = ".gh-issue-dependency-cache"
	CacheDuration = 5 * time.Minute // Cache for 5 minutes
)

// Repository Context Detection
//
// These functions handle repository context detection following gh-sub-issue patterns.
// They provide a consistent way to determine which repository to work with based on
// user input, current directory, and command-line flags.

// GetCurrentRepo gets the current repository context using the GitHub CLI.
// It uses 'gh repo view' to determine the repository based on the current working directory.
// This function requires that the user is in a directory associated with a GitHub repository
// and that they have authenticated with the GitHub CLI.
//
// Returns the repository owner and name, or an error if the repository cannot be determined.
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

// GitHub API Integration for Issue Dependencies
//
// These functions implement the GitHub API integration for fetching issue dependency
// relationships using the go-gh/v2 library. They handle parallel API calls, error
// handling, and data transformation.

// fetchIssueDetails retrieves issue details from the GitHub API
func fetchIssueDetails(ctx context.Context, client *api.RESTClient, owner, repo string, issueNumber int) (*Issue, error) {
	// API endpoint for issue details
	endpoint := fmt.Sprintf("repos/%s/%s/issues/%d", owner, repo, issueNumber)
	
	var issue Issue
	err := client.Get(endpoint, &issue)
	if err != nil {
		// Handle specific error types
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return nil, NewIssueNotFoundError(fmt.Sprintf("%s/%s", owner, repo), issueNumber)
		}
		if strings.Contains(strings.ToLower(err.Error()), "forbidden") {
			return nil, WrapPermissionError(fmt.Sprintf("%s/%s", owner, repo), err)
		}
		if strings.Contains(strings.ToLower(err.Error()), "unauthorized") {
			return nil, WrapAuthError(err)
		}
		if strings.Contains(strings.ToLower(err.Error()), "rate limit") {
			return nil, WrapAPIError(429, err)
		}
		
		return nil, WrapInternalError("fetching issue details", err)
	}
	
	// Add repository information for cross-repo support
	issue.Repository = fmt.Sprintf("%s/%s", owner, repo)
	
	return &issue, nil
}

// fetchDependencyRelationships retrieves dependency relationships from GitHub API
func fetchDependencyRelationships(ctx context.Context, client *api.RESTClient, owner, repo string, issueNumber int, relationType string) ([]DependencyRelation, error) {
	// API endpoint for dependency relationships
	endpoint := fmt.Sprintf("repos/%s/%s/issues/%d/dependencies/%s", owner, repo, issueNumber, relationType)
	
	var relations []struct {
		Issue Issue `json:"issue"`
	}
	
	err := client.Get(endpoint, &relations)
	if err != nil {
		// Handle specific error types
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			// Issue doesn't exist or no dependencies - return empty slice
			return []DependencyRelation{}, nil
		}
		if strings.Contains(strings.ToLower(err.Error()), "forbidden") {
			return nil, WrapPermissionError(fmt.Sprintf("%s/%s", owner, repo), err)
		}
		if strings.Contains(strings.ToLower(err.Error()), "unauthorized") {
			return nil, WrapAuthError(err)
		}
		if strings.Contains(strings.ToLower(err.Error()), "rate limit") {
			return nil, WrapAPIError(429, err)
		}
		
		return nil, WrapInternalError(fmt.Sprintf("fetching %s dependencies", relationType), err)
	}
	
	// Transform to DependencyRelation objects
	var dependencies []DependencyRelation
	for _, rel := range relations {
		// Extract repository from issue HTML URL if available
		repoName := fmt.Sprintf("%s/%s", owner, repo) // Default to current repo
		if rel.Issue.HTMLURL != "" {
			if repoFromURL := extractRepoFromURL(rel.Issue.HTMLURL); repoFromURL != "" {
				repoName = repoFromURL
			}
		}
		
		dependencies = append(dependencies, DependencyRelation{
			Issue:      rel.Issue,
			Type:       relationType,
			Repository: repoName,
		})
	}
	
	return dependencies, nil
}

// extractRepoFromURL extracts repository name from GitHub issue URL
func extractRepoFromURL(url string) string {
	// Extract repo from URL format: https://github.com/owner/repo/issues/123
	if !strings.HasPrefix(url, "https://github.com/") {
		return ""
	}
	
	parts := strings.Split(strings.TrimPrefix(url, "https://github.com/"), "/")
	if len(parts) < 2 {
		return ""
	}
	
	return parts[0] + "/" + parts[1]
}

// fetchDependencies retrieves all dependency data for an issue using parallel API calls
func fetchDependencies(ctx context.Context, owner, repo string, issueNumber int) (*DependencyData, error) {
	// Create GitHub API client
	client, err := api.DefaultRESTClient()
	if err != nil {
		return nil, WrapInternalError("creating GitHub API client", err)
	}
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	// Channel for collecting results
	type fetchResult struct {
		sourceIssue *Issue
		blockedBy   []DependencyRelation
		blocking    []DependencyRelation
		err         error
	}
	
	resultChan := make(chan fetchResult, 3)
	var wg sync.WaitGroup
	
	// Fetch source issue details
	wg.Add(1)
	go func() {
		defer wg.Done()
		issue, err := fetchIssueDetails(ctx, client, owner, repo, issueNumber)
		resultChan <- fetchResult{sourceIssue: issue, err: err}
	}()
	
	// Fetch blocked_by relationships
	wg.Add(1)
	go func() {
		defer wg.Done()
		relations, err := fetchDependencyRelationships(ctx, client, owner, repo, issueNumber, "blocked_by")
		resultChan <- fetchResult{blockedBy: relations, err: err}
	}()
	
	// Fetch blocking relationships  
	wg.Add(1)
	go func() {
		defer wg.Done()
		relations, err := fetchDependencyRelationships(ctx, client, owner, repo, issueNumber, "blocking")
		resultChan <- fetchResult{blocking: relations, err: err}
	}()
	
	// Wait for all goroutines to complete
	wg.Wait()
	close(resultChan)
	
	// Collect results
	var data DependencyData
	data.FetchedAt = time.Now()
	
	for result := range resultChan {
		if result.err != nil {
			return nil, result.err
		}
		
		if result.sourceIssue != nil {
			data.SourceIssue = *result.sourceIssue
		}
		if result.blockedBy != nil {
			data.BlockedBy = result.blockedBy
		}
		if result.blocking != nil {
			data.Blocking = result.blocking
		}
	}
	
	// Calculate total count
	data.TotalCount = len(data.BlockedBy) + len(data.Blocking)
	
	return &data, nil
}

// FetchIssueDependencies is the main exported function for retrieving dependency data
func FetchIssueDependencies(ctx context.Context, owner, repo string, issueNumber int) (*DependencyData, error) {
	// Validate inputs
	if owner == "" || repo == "" {
		return nil, NewEmptyValueError("repository owner or name")
	}
	if issueNumber <= 0 {
		return nil, NewIssueNumberValidationError(strconv.Itoa(issueNumber))
	}
	
	// Try to get from cache first
	cacheKey := getCacheKey(owner, repo, issueNumber)
	if data, found := getFromCache(cacheKey); found {
		return data, nil
	}
	
	// Verify GitHub CLI authentication
	if err := SetupGitHubClient(); err != nil {
		return nil, err
	}
	
	// Validate repository access
	if err := ValidateRepoAccess(owner, repo); err != nil {
		return nil, err
	}
	
	// Fetch dependency data
	data, err := fetchDependencies(ctx, owner, repo, issueNumber)
	if err != nil {
		return nil, err
	}
	
	// Cache the result
	saveToCache(cacheKey, data)
	
	return data, nil
}

// getCacheKey generates a unique cache key for the request
func getCacheKey(owner, repo string, issueNumber int) string {
	key := fmt.Sprintf("%s/%s#%d", owner, repo, issueNumber)
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("%x", hash)
}

// getCacheDir returns the cache directory path
func getCacheDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return CacheDir // fallback to relative path
	}
	return filepath.Join(homeDir, CacheDir)
}

// getFromCache attempts to retrieve data from cache
func getFromCache(key string) (*DependencyData, bool) {
	cacheDir := getCacheDir()
	cachePath := filepath.Join(cacheDir, key+".json")
	
	// Check if cache file exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return nil, false
	}
	
	// Read cache file
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, false
	}
	
	// Parse cache entry
	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}
	
	// Check if cache entry has expired
	if time.Now().After(entry.ExpiresAt) {
		// Remove expired cache file
		os.Remove(cachePath)
		return nil, false
	}
	
	return &entry.Data, true
}

// saveToCache stores data in cache
func saveToCache(key string, data *DependencyData) {
	cacheDir := getCacheDir()
	
	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return // fail silently
	}
	
	// Create cache entry
	entry := CacheEntry{
		Data:      *data,
		ExpiresAt: time.Now().Add(CacheDuration),
	}
	
	// Marshal to JSON
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return // fail silently
	}
	
	// Write to cache file
	cachePath := filepath.Join(cacheDir, key+".json")
	os.WriteFile(cachePath, jsonData, 0644)
}

// CleanExpiredCache removes expired cache entries
func CleanExpiredCache() error {
	cacheDir := getCacheDir()
	
	// Check if cache directory exists
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return nil // no cache to clean
	}
	
	// Read cache directory
	files, err := os.ReadDir(cacheDir)
	if err != nil {
		return err
	}
	
	now := time.Now()
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		
		cachePath := filepath.Join(cacheDir, file.Name())
		
		// Read cache file
		data, err := os.ReadFile(cachePath)
		if err != nil {
			continue
		}
		
		// Parse cache entry
		var entry CacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			// Remove malformed cache files
			os.Remove(cachePath)
			continue
		}
		
		// Remove expired entries
		if now.After(entry.ExpiresAt) {
			os.Remove(cachePath)
		}
	}
	
	return nil
}