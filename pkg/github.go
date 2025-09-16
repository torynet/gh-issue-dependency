// Package pkg provides shared utilities and types for the gh-issue-dependency extension.
//
// This package contains GitHub API integration, error handling, repository context
// detection, and other common functionality used across all commands.
package pkg

import (
	"context"
	"crypto/sha256"
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
	Number     int            `json:"number"`
	Title      string         `json:"title"`
	State      string         `json:"state"`
	Assignees  []User         `json:"assignees"`
	Labels     []Label        `json:"labels"`
	HTMLURL    string         `json:"html_url"`
	Repository RepositoryInfo `json:"repository,omitempty"` // Repository object from GitHub API
}

// RepositoryInfo represents repository information from GitHub API
type RepositoryInfo struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
	Owner    struct {
		Login string `json:"login"`
	} `json:"owner"`
}

// String returns the string representation of the repository (FullName)
func (r RepositoryInfo) String() string {
	return r.FullName
}

// IsEmpty returns true if this is an empty/zero repository info
func (r RepositoryInfo) IsEmpty() bool {
	return r.FullName == ""
}

// DependencyRelation represents a relationship between issues
type DependencyRelation struct {
	Issue      Issue  `json:"issue"`
	Type       string `json:"type"`       // "blocked_by" or "blocks"
	Repository string `json:"repository"` // Repository of the related issue
}

// DependencyData contains all dependency information for an issue
type DependencyData struct {
	SourceIssue            Issue                `json:"source_issue"`
	BlockedBy              []DependencyRelation `json:"blocked_by"`
	Blocking               []DependencyRelation `json:"blocking"`
	FetchedAt              time.Time            `json:"fetched_at"`
	TotalCount             int                  `json:"total_count"`
	OriginalBlockedByCount int                  `json:"original_blocked_by_count,omitempty"`
	OriginalBlockingCount  int                  `json:"original_blocking_count,omitempty"`
}

// CacheEntry represents a cached dependency data entry
type CacheEntry struct {
	Data      DependencyData `json:"data"`
	ExpiresAt time.Time      `json:"expires_at"`
}

// Cache configuration
const (
	CacheDir      = ".gh-issue-dependency-cache"
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
	// Validate repository name format to prevent command injection
	repoName := fmt.Sprintf("%s/%s", owner, repo)
	if !regexp.MustCompile(`^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`).MatchString(repoName) {
		return fmt.Errorf("invalid repository name format: %s", repoName)
	}
	cmd := exec.Command("gh", "repo", "view", repoName, "--json", "id") // #nosec G204 -- repoName validated with strict regex
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
	if strings.HasPrefix(issueRef, "https://") {
		if !strings.HasPrefix(issueRef, "https://github.com/") {
			return "", "", fmt.Errorf("unsupported URL: only GitHub URLs are supported, got: %s", issueRef)
		}
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
	issue.Repository = RepositoryInfo{
		Name:     repo,
		FullName: fmt.Sprintf("%s/%s", owner, repo),
		Owner: struct {
			Login string `json:"login"`
		}{Login: owner},
	}

	return &issue, nil
}

// fetchDependencyRelationships retrieves dependency relationships from GitHub API
func fetchDependencyRelationships(ctx context.Context, client *api.RESTClient, owner, repo string, issueNumber int, relationType string) ([]DependencyRelation, error) {
	// API endpoint for dependency relationships
	endpoint := fmt.Sprintf("repos/%s/%s/issues/%d/dependencies/%s", owner, repo, issueNumber, relationType)

	// Use generic interface to handle GitHub's API response format
	var rawData interface{}

	err := client.Get(endpoint, &rawData)
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

	// Now parse the data into our Issue structs
	var relations []Issue
	if rawDataBytes, err := json.Marshal(rawData); err == nil {
		if err := json.Unmarshal(rawDataBytes, &relations); err != nil {
			return nil, WrapInternalError(fmt.Sprintf("parsing %s dependencies", relationType), err)
		}
	}

	// Transform to DependencyRelation objects
	var dependencies []DependencyRelation
	for _, rel := range relations {
		// Add validation to catch unmarshaling issues
		if rel.Number == 0 {
			continue // Skip zero-value issues that indicate unmarshaling problems
		}

		// Extract repository name - use the repository field if available, otherwise use the current repo
		repoName := fmt.Sprintf("%s/%s", owner, repo) // Default to current repo
		if rel.Repository.FullName != "" {
			repoName = rel.Repository.FullName
		} else if rel.HTMLURL != "" {
			// Fallback: extract from HTML URL
			if repoFromURL := extractRepoFromURL(rel.HTMLURL); repoFromURL != "" {
				repoName = repoFromURL
			}
		}

		dependencies = append(dependencies, DependencyRelation{
			Issue:      rel,
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
	hash := sha256.Sum256([]byte(key))
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

	// Validate path to prevent directory traversal
	if !strings.HasPrefix(cachePath, cacheDir) {
		return nil, false
	}

	// Check if cache file exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return nil, false
	}

	// Read cache file
	data, err := os.ReadFile(cachePath) // #nosec G304 -- cachePath validated against directory traversal
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
		_ = os.Remove(cachePath) // Ignore cleanup errors
		return nil, false
	}

	return &entry.Data, true
}

// saveToCache stores data in cache
func saveToCache(key string, data *DependencyData) {
	cacheDir := getCacheDir()

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0750); err != nil {
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
	if err := os.WriteFile(cachePath, jsonData, 0600); err != nil {
		// Log error but don't fail the main operation
		fmt.Fprintf(os.Stderr, "Warning: failed to write cache file: %v\n", err)
	}
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

		// Validate path to prevent directory traversal
		if !strings.HasPrefix(cachePath, cacheDir) {
			continue
		}

		// Read cache file
		data, err := os.ReadFile(cachePath) // #nosec G304 -- cachePath validated against directory traversal
		if err != nil {
			continue
		}

		// Parse cache entry
		var entry CacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			// Remove malformed cache files
			_ = os.Remove(cachePath) // Ignore cleanup errors
			continue
		}

		// Remove expired entries
		if now.After(entry.ExpiresAt) {
			_ = os.Remove(cachePath) // Ignore cleanup errors
		}
	}

	return nil
}

// GitHub API Integration for Dependency Removal
//
// These functions implement the GitHub API integration for deleting issue dependency
// relationships using DELETE operations. They handle retry logic, error handling,
// and integration with the validation system.

// DependencyRemover provides GitHub API integration for removing dependency relationships.
// It handles DELETE operations, error processing, retry logic, and success confirmation.
type DependencyRemover struct {
	client    *api.RESTClient
	validator *RemovalValidator
}

// NewDependencyRemover creates a new dependency remover with GitHub API client
func NewDependencyRemover() (*DependencyRemover, error) {
	// Verify GitHub CLI authentication
	if err := SetupGitHubClient(); err != nil {
		return nil, err
	}

	// Create GitHub API client
	client, err := api.DefaultRESTClient()
	if err != nil {
		return nil, WrapInternalError("creating GitHub API client for removal", err)
	}

	// Create validator for removal operations
	validator, err := NewRemovalValidator()
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %w", err)
	}

	return &DependencyRemover{
		client:    client,
		validator: validator,
	}, nil
}

// RemoveRelationship removes a single dependency relationship between two issues
func (r *DependencyRemover) RemoveRelationship(source, target IssueRef, relType string, opts RemoveOptions) error {
	// 1. Run comprehensive validation pipeline
	if err := r.validator.ValidateRemoval(source, target, relType); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 2. Handle dry run mode
	if opts.DryRun {
		return r.showDryRunPreview(source, target, relType)
	}

	// 3. Get user confirmation (unless --force is specified)
	if !opts.Force {
		confirmed, err := r.requestConfirmation(source, target, relType)
		if err != nil {
			return fmt.Errorf("confirmation failed: %w", err)
		}
		if !confirmed {
			return NewAppError(
				ErrorTypeValidation,
				"Dependency removal cancelled by user",
				nil,
			).WithSuggestion("Use --force to skip confirmation prompts")
		}
	}

	// 4. Execute deletion with retry logic
	if err := r.deleteRelationshipWithRetry(source, target, relType); err != nil {
		return fmt.Errorf("deletion failed: %w", err)
	}

	// 5. Show success confirmation
	return r.showSuccessMessage(source, target, relType)
}

// RemoveBatchRelationships removes multiple dependency relationships in batch
func (r *DependencyRemover) RemoveBatchRelationships(source IssueRef, targets []IssueRef, relType string, opts RemoveOptions) error {
	// 1. Run batch validation
	if err := r.validator.ValidateBatchRemoval(source, targets, relType); err != nil {
		return fmt.Errorf("batch validation failed: %w", err)
	}

	// 2. Handle dry run mode
	if opts.DryRun {
		return r.showBatchDryRunPreview(source, targets, relType)
	}

	// 3. Get user confirmation for batch operation
	if !opts.Force {
		confirmed, err := r.requestBatchConfirmation(source, targets, relType)
		if err != nil {
			return fmt.Errorf("batch confirmation failed: %w", err)
		}
		if !confirmed {
			return NewAppError(
				ErrorTypeValidation,
				"Batch dependency removal cancelled by user",
				nil,
			).WithSuggestion("Use --force to skip confirmation prompts")
		}
	}

	// 4. Execute batch deletion
	return r.executeBatchDeletion(source, targets, relType)
}

// deleteRelationshipWithRetry performs the actual DELETE operation with retry logic
func (r *DependencyRemover) deleteRelationshipWithRetry(source, target IssueRef, relType string) error {
	maxRetries := 3
	baseDelay := 1 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := r.deleteRelationship(source, target, relType)
		if err == nil {
			return nil // Success
		}

		// Check if error is retryable
		if !r.isRetryableError(err) {
			return err // Don't retry for non-retryable errors
		}

		// Don't retry on last attempt
		if attempt == maxRetries {
			return fmt.Errorf("deletion failed after %d attempts: %w", maxRetries, err)
		}

		// Exponential backoff delay
		delay := time.Duration(attempt) * baseDelay
		time.Sleep(delay)
	}

	return nil
}

// deleteRelationship performs the actual DELETE API call
func (r *DependencyRemover) deleteRelationship(source, target IssueRef, relType string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// First, get the relationship ID by finding the specific relationship
	relationshipID, err := r.findRelationshipID(ctx, source, target, relType)
	if err != nil {
		return fmt.Errorf("failed to find relationship ID: %w", err)
	}

	// Construct the DELETE endpoint
	endpoint := fmt.Sprintf("repos/%s/%s/issues/%d/dependencies/%s",
		source.Owner, source.Repo, source.Number, relationshipID)

	// Execute DELETE request
	err = r.client.Delete(endpoint, nil)
	if err != nil {
		return r.handleDeleteError(err, source, target, relType)
	}

	return nil
}

// findRelationshipID finds the specific relationship ID for deletion
func (r *DependencyRemover) findRelationshipID(ctx context.Context, source, target IssueRef, relType string) (string, error) {
	// Get current dependencies to find the relationship ID
	dependencies, err := r.validator.fetchIssueDependencies(ctx, source)
	if err != nil {
		return "", fmt.Errorf("failed to fetch dependencies: %w", err)
	}

	// Find the specific relationship
	var relations []DependencyRelation
	switch relType {
	case "blocked-by":
		relations = dependencies.BlockedBy
	case "blocks":
		relations = dependencies.Blocking
	default:
		return "", NewAppError(
			ErrorTypeValidation,
			fmt.Sprintf("Invalid relationship type: %s", relType),
			nil,
		).WithSuggestion("Use either 'blocked-by' or 'blocks'")
	}

	// Find matching relationship
	for _, relation := range relations {
		if r.matchesTarget(relation, target) {
			// For GitHub's API, we typically use a combination of repo and issue number
			// The relationship ID format depends on GitHub's implementation
			return fmt.Sprintf("%s#%d", target.FullName, target.Number), nil
		}
	}

	return "", NewAppError(
		ErrorTypeIssue,
		fmt.Sprintf("Relationship not found: %s %s %s", source.String(), relType, target.String()),
		nil,
	).WithSuggestion("Use 'gh issue-dependency list' to see current dependencies")
}

// matchesTarget checks if a dependency relation matches the target issue
func (r *DependencyRemover) matchesTarget(relation DependencyRelation, target IssueRef) bool {
	if relation.Issue.Number != target.Number {
		return false
	}

	targetRepo := target.FullName
	if targetRepo == "" {
		targetRepo = fmt.Sprintf("%s/%s", target.Owner, target.Repo)
	}

	return relation.Repository == targetRepo || relation.Issue.Repository.FullName == targetRepo
}

// handleDeleteError processes and categorizes deletion errors
func (r *DependencyRemover) handleDeleteError(err error, source, target IssueRef, relType string) error {
	errMsg := strings.ToLower(err.Error())

	// Authentication errors
	if strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "401") {
		return WrapAuthError(err).WithSuggestion("Run 'gh auth login' to authenticate")
	}

	// Permission errors
	if strings.Contains(errMsg, "forbidden") || strings.Contains(errMsg, "403") {
		repoName := fmt.Sprintf("%s/%s", source.Owner, source.Repo)
		return NewPermissionDeniedError("remove dependencies", repoName).WithSuggestion(
			"You need write or maintain permissions to modify dependencies")
	}

	// Not found errors - relationship may have been removed already
	if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "404") {
		return NewAppError(
			ErrorTypeIssue,
			fmt.Sprintf("Dependency relationship no longer exists: %s %s %s",
				source.String(), relType, target.String()),
			err,
		).WithSuggestion("The relationship may have been removed by another process")
	}

	// Rate limiting
	if strings.Contains(errMsg, "rate limit") || strings.Contains(errMsg, "429") {
		return WrapAPIError(429, err)
	}

	// Network errors
	if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "connection") {
		return WrapNetworkError(err)
	}

	// Server errors
	if strings.Contains(errMsg, "500") || strings.Contains(errMsg, "502") || strings.Contains(errMsg, "503") {
		return WrapAPIError(500, err)
	}

	// Default internal error
	return WrapInternalError("removing dependency relationship", err)
}

// isRetryableError determines if an error should trigger a retry
func (r *DependencyRemover) isRetryableError(err error) bool {
	// Check for retryable error types
	if IsErrorType(err, ErrorTypeNetwork) {
		return true
	}
	if IsErrorType(err, ErrorTypeAPI) {
		// Rate limits and server errors are retryable
		errMsg := strings.ToLower(err.Error())
		return strings.Contains(errMsg, "rate limit") ||
			strings.Contains(errMsg, "500") ||
			strings.Contains(errMsg, "502") ||
			strings.Contains(errMsg, "503")
	}

	return false
}

// executeBatchDeletion performs batch deletion of multiple relationships
func (r *DependencyRemover) executeBatchDeletion(source IssueRef, targets []IssueRef, relType string) error {
	var errors []string
	successCount := 0

	for _, target := range targets {
		err := r.deleteRelationshipWithRetry(source, target, relType)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", target.String(), err))
		} else {
			successCount++
		}
	}

	// Report results
	if len(errors) > 0 {
		return NewAppError(
			ErrorTypeAPI,
			fmt.Sprintf("Batch removal partially failed: %d succeeded, %d failed",
				successCount, len(errors)),
			nil,
		).WithContext("errors", strings.Join(errors, "; ")).WithSuggestion(
			"Review the errors and retry failed operations individually")
	}

	return r.showBatchSuccessMessage(source, targets, relType)
}

// User Interface and Confirmation Functions
//
// These functions handle user interaction, confirmation prompts, dry run previews,
// and success/failure reporting for dependency removal operations.

// showDryRunPreview displays what would be removed in dry run mode
func (r *DependencyRemover) showDryRunPreview(source, target IssueRef, relType string) error {
	fmt.Printf("Dry run: dependency removal preview\n\n")

	var relationshipDescription string
	switch relType {
	case "blocked-by":
		relationshipDescription = fmt.Sprintf("blocked-by relationship: %s ← %s", source.String(), target.String())
	case "blocks":
		relationshipDescription = fmt.Sprintf("blocks relationship: %s → %s", source.String(), target.String())
	}

	fmt.Printf("Would remove:\n")
	fmt.Printf("  ❌ %s\n", relationshipDescription)
	fmt.Printf("\nNo changes made. Use --force to skip confirmation or remove --dry-run to execute.\n")

	return nil
}

// showBatchDryRunPreview displays batch removal preview
func (r *DependencyRemover) showBatchDryRunPreview(source IssueRef, targets []IssueRef, relType string) error {
	fmt.Printf("Dry run: batch dependency removal preview\n\n")
	fmt.Printf("Would remove %d relationships:\n", len(targets))

	for _, target := range targets {
		var relationshipDescription string
		switch relType {
		case "blocked-by":
			relationshipDescription = fmt.Sprintf("blocked-by relationship: %s ← %s", source.String(), target.String())
		case "blocks":
			relationshipDescription = fmt.Sprintf("blocks relationship: %s → %s", source.String(), target.String())
		}
		fmt.Printf("  ❌ %s\n", relationshipDescription)
	}

	fmt.Printf("\nNo changes made. Use --force to skip confirmation or remove --dry-run to execute.\n")

	return nil
}

// requestConfirmation prompts user for confirmation before removing a relationship
func (r *DependencyRemover) requestConfirmation(source, target IssueRef, relType string) (bool, error) {
	// Get issue details for better confirmation prompt
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sourceIssue, err := fetchIssueDetails(ctx, r.client, source.Owner, source.Repo, source.Number)
	if err != nil {
		// Fall back to basic confirmation if we can't get details
		return r.requestBasicConfirmation(source, target, relType)
	}

	targetIssue, err := fetchIssueDetails(ctx, r.client, target.Owner, target.Repo, target.Number)
	if err != nil {
		// Fall back to basic confirmation if we can't get details
		return r.requestBasicConfirmation(source, target, relType)
	}

	fmt.Printf("Remove dependency relationship?\n")
	fmt.Printf("  Source: %s - %s\n", source.String(), sourceIssue.Title)
	fmt.Printf("  Target: %s - %s\n", target.String(), targetIssue.Title)
	fmt.Printf("  Type: %s\n\n", relType)

	var relationshipDescription string
	switch relType {
	case "blocked-by":
		relationshipDescription = fmt.Sprintf("This will remove the \"%s\" relationship between these issues.", relType)
	case "blocks":
		relationshipDescription = fmt.Sprintf("This will remove the \"%s\" relationship between these issues.", relType)
	}

	fmt.Printf("%s\n", relationshipDescription)
	fmt.Printf("Continue? (y/N): ")

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		// Default to "no" on input error for safety
		response = "n"
	} else {
		response = strings.ToLower(strings.TrimSpace(response))
	}
	return response == "y" || response == "yes", nil
}

// requestBasicConfirmation provides a basic confirmation prompt when issue details aren't available
func (r *DependencyRemover) requestBasicConfirmation(source, target IssueRef, relType string) (bool, error) {
	fmt.Printf("Remove %s dependency relationship between %s and %s?\n", relType, source.String(), target.String())
	fmt.Printf("Continue? (y/N): ")

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		// Default to "no" on input error for safety
		response = "n"
	} else {
		response = strings.ToLower(strings.TrimSpace(response))
	}
	return response == "y" || response == "yes", nil
}

// requestBatchConfirmation prompts user for confirmation before batch removal
func (r *DependencyRemover) requestBatchConfirmation(source IssueRef, targets []IssueRef, relType string) (bool, error) {
	fmt.Printf("Remove %d dependency relationships?\n", len(targets))
	fmt.Printf("  Source: %s\n", source.String())
	fmt.Printf("  Type: %s\n", relType)
	fmt.Printf("  Targets:\n")

	for _, target := range targets {
		fmt.Printf("    - %s\n", target.String())
	}

	fmt.Printf("\nThis will remove %d dependency relationships.\n", len(targets))
	fmt.Printf("Continue? (y/N): ")

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		// Default to "no" on input error for safety
		response = "n"
	} else {
		response = strings.ToLower(strings.TrimSpace(response))
	}
	return response == "y" || response == "yes", nil
}

// showSuccessMessage displays success confirmation after removing a relationship
func (r *DependencyRemover) showSuccessMessage(source, target IssueRef, relType string) error {
	var relationshipSymbol string
	switch relType {
	case "blocked-by":
		relationshipSymbol = "←"
	case "blocks":
		relationshipSymbol = "→"
	}

	fmt.Printf("✅ Removed %s relationship: %s %s %s\n\n",
		relType, source.String(), relationshipSymbol, target.String())
	fmt.Printf("Dependency removed successfully.\n")

	return nil
}

// showBatchSuccessMessage displays success confirmation after batch removal
func (r *DependencyRemover) showBatchSuccessMessage(source IssueRef, targets []IssueRef, relType string) error {
	var relationshipSymbol string
	switch relType {
	case "blocked-by":
		relationshipSymbol = "←"
	case "blocks":
		relationshipSymbol = "→"
	}

	fmt.Printf("✅ Removed %d %s relationships:\n", len(targets), relType)
	for _, target := range targets {
		fmt.Printf("  %s %s %s\n", source.String(), relationshipSymbol, target.String())
	}
	fmt.Printf("\nBatch dependency removal completed successfully.\n")

	return nil
}

// Advanced DELETE Operations with Cross-Repository Support
//
// These functions extend the basic DELETE operations to handle cross-repository
// dependency deletion and provide enhanced error handling for complex scenarios.

// RemoveCrossRepositoryRelationship handles dependency removal across different repositories
func (r *DependencyRemover) RemoveCrossRepositoryRelationship(source, target IssueRef, relType string, opts RemoveOptions) error {
	// Cross-repository relationships require additional validation
	if err := r.validateCrossRepositoryPermissions(source, target); err != nil {
		return fmt.Errorf("cross-repository validation failed: %w", err)
	}

	// Use the standard removal process
	return r.RemoveRelationship(source, target, relType, opts)
}

// validateCrossRepositoryPermissions ensures user has permissions in both repositories
func (r *DependencyRemover) validateCrossRepositoryPermissions(source, target IssueRef) error {
	// Validate source repository permissions
	if err := ValidateRepoAccess(source.Owner, source.Repo); err != nil {
		return fmt.Errorf("source repository access failed: %w", err)
	}

	// Validate target repository permissions (for cross-repo dependencies)
	if source.Owner != target.Owner || source.Repo != target.Repo {
		if err := ValidateRepoAccess(target.Owner, target.Repo); err != nil {
			return fmt.Errorf("target repository access failed: %w", err)
		}
	}

	return nil
}

// RemoveAllRelationships removes all dependency relationships for an issue
func (r *DependencyRemover) RemoveAllRelationships(issue IssueRef, opts RemoveOptions) error {
	// Get current dependencies
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dependencies, err := r.validator.fetchIssueDependencies(ctx, issue)
	if err != nil {
		return fmt.Errorf("failed to fetch dependencies for removal: %w", err)
	}

	if len(dependencies.BlockedBy) == 0 && len(dependencies.Blocking) == 0 {
		return NewAppError(
			ErrorTypeIssue,
			fmt.Sprintf("No dependency relationships found for %s", issue.String()),
			nil,
		).WithSuggestion("Use 'gh issue-dependency list' to see current dependencies")
	}

	// Collect all targets
	var blockedByTargets []IssueRef
	var blockingTargets []IssueRef

	for _, relation := range dependencies.BlockedBy {
		target := r.dependencyRelationToIssueRef(relation)
		blockedByTargets = append(blockedByTargets, target)
	}

	for _, relation := range dependencies.Blocking {
		target := r.dependencyRelationToIssueRef(relation)
		blockingTargets = append(blockingTargets, target)
	}

	// Remove blocked-by relationships
	if len(blockedByTargets) > 0 {
		if err := r.RemoveBatchRelationships(issue, blockedByTargets, "blocked-by", opts); err != nil {
			return fmt.Errorf("failed to remove blocked-by relationships: %w", err)
		}
	}

	// Remove blocking relationships
	if len(blockingTargets) > 0 {
		if err := r.RemoveBatchRelationships(issue, blockingTargets, "blocks", opts); err != nil {
			return fmt.Errorf("failed to remove blocks relationships: %w", err)
		}
	}

	fmt.Printf("✅ Removed all dependency relationships for %s\n", issue.String())
	fmt.Printf("  - %d blocked-by relationships removed\n", len(blockedByTargets))
	fmt.Printf("  - %d blocks relationships removed\n\n", len(blockingTargets))
	fmt.Printf("All dependencies cleared successfully.\n")

	return nil
}

// dependencyRelationToIssueRef converts a DependencyRelation to an IssueRef
func (r *DependencyRemover) dependencyRelationToIssueRef(relation DependencyRelation) IssueRef {
	// Parse the repository from the relation
	repoParts := strings.Split(relation.Repository, "/")
	if len(repoParts) != 2 {
		// Fallback - extract from issue repository field
		if relation.Issue.Repository.FullName != "" {
			repoParts = strings.Split(relation.Issue.Repository.FullName, "/")
		} else {
			// Last resort - use empty values, which will cause validation errors
			repoParts = []string{"", ""}
		}
	}

	owner := ""
	repo := ""
	if len(repoParts) == 2 {
		owner = repoParts[0]
		repo = repoParts[1]
	}

	return IssueRef{
		Owner:    owner,
		Repo:     repo,
		Number:   relation.Issue.Number,
		FullName: relation.Repository,
	}
}
