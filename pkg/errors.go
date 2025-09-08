package pkg

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Error Handling and User-Friendly Error Messages
//
// This file provides structured error handling with contextual information
// and user-friendly error messages. It categorizes errors by type and provides
// specific suggestions for resolution.

// ErrorType represents the category of error for proper handling.
// This allows the CLI to determine appropriate exit codes and error formatting.
type ErrorType string

const (
	ErrorTypeAuthentication ErrorType = "authentication"
	ErrorTypePermission     ErrorType = "permission"
	ErrorTypeNetwork        ErrorType = "network"
	ErrorTypeValidation     ErrorType = "validation"
	ErrorTypeAPI            ErrorType = "api"
	ErrorTypeRepository     ErrorType = "repository"
	ErrorTypeIssue          ErrorType = "issue"
	ErrorTypeInternal       ErrorType = "internal"
)

// AppError represents a structured error with context and user guidance.
// This is the primary error type used throughout the application to provide
// consistent, user-friendly error messages with actionable suggestions.
type AppError struct {
	Type        ErrorType         // Category of error for exit code determination
	Message     string            // User-facing error message
	Cause       error             // Underlying error that caused this error
	Context     map[string]string // Additional context information (repository, issue, etc.)
	Suggestions []string          // Actionable suggestions for resolving the error
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewAppError creates a new structured application error
func NewAppError(errType ErrorType, message string, cause error) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Cause:   cause,
		Context: make(map[string]string),
	}
}

// WithContext adds contextual information to an error
func (e *AppError) WithContext(key, value string) *AppError {
	e.Context[key] = value
	return e
}

// WithSuggestion adds a recovery suggestion to an error
func (e *AppError) WithSuggestion(suggestion string) *AppError {
	e.Suggestions = append(e.Suggestions, suggestion)
	return e
}

// Authentication Errors
func WrapAuthError(err error) *AppError {
	return NewAppError(
		ErrorTypeAuthentication,
		"Authentication required to access GitHub",
		err,
	).WithSuggestion("Run 'gh auth login' to authenticate with GitHub")
}

func NewAuthTokenError() *AppError {
	return NewAppError(
		ErrorTypeAuthentication,
		"Invalid or expired GitHub token",
		nil,
	).WithSuggestion("Run 'gh auth login' to refresh your authentication")
}

// Permission Errors
func WrapPermissionError(repo string, err error) *AppError {
	return NewAppError(
		ErrorTypePermission,
		fmt.Sprintf("Insufficient permissions to access %s", repo),
		err,
	).WithContext("repository", repo).
		WithSuggestion("Ensure you have at least triage permissions for this repository").
		WithSuggestion("Contact the repository owner to request access")
}

func NewPermissionDeniedError(operation, repo string) *AppError {
	return NewAppError(
		ErrorTypePermission,
		fmt.Sprintf("Permission denied: cannot %s in %s", operation, repo),
		nil,
	).WithContext("operation", operation).
		WithContext("repository", repo).
		WithSuggestion("Verify you have the required permissions for this operation")
}

// Network Errors
func WrapNetworkError(err error) *AppError {
	return NewAppError(
		ErrorTypeNetwork,
		"Network error occurred while connecting to GitHub",
		err,
	).WithSuggestion("Check your internet connection and retry").
		WithSuggestion("Verify GitHub's service status at https://www.githubstatus.com/")
}

func NewTimeoutError(operation string) *AppError {
	return NewAppError(
		ErrorTypeNetwork,
		fmt.Sprintf("Request timed out while %s", operation),
		nil,
	).WithContext("operation", operation).
		WithSuggestion("Retry the operation").
		WithSuggestion("Check your network connection")
}

// Validation Errors
func WrapValidationError(field, value string, err error) *AppError {
	return NewAppError(
		ErrorTypeValidation,
		fmt.Sprintf("Invalid %s: %s", field, value),
		err,
	).WithContext("field", field).
		WithContext("value", value)
}

func NewIssueNumberValidationError(value string) *AppError {
	return NewAppError(
		ErrorTypeValidation,
		fmt.Sprintf("Invalid issue number format: %s (expected number or GitHub URL)", value),
		nil,
	).WithContext("input", value).
		WithSuggestion("Use a numeric issue number (e.g., 123)").
		WithSuggestion("Use a GitHub issue URL (e.g., https://github.com/owner/repo/issues/123)").
		WithSuggestion("Use owner/repo#123 format for cross-repository references")
}

func NewRepositoryFormatError(value string) *AppError {
	return NewAppError(
		ErrorTypeValidation,
		fmt.Sprintf("Invalid repository format: %s", value),
		nil,
	).WithContext("input", value).
		WithSuggestion("Use OWNER/REPO format (e.g., octocat/Hello-World)").
		WithSuggestion("Use full GitHub URL (e.g., https://github.com/octocat/Hello-World)")
}

func NewEmptyValueError(field string) *AppError {
	return NewAppError(
		ErrorTypeValidation,
		fmt.Sprintf("%s cannot be empty", field),
		nil,
	).WithContext("field", field)
}

// API Errors
func WrapAPIError(statusCode int, err error) *AppError {
	switch statusCode {
	case http.StatusUnauthorized:
		return WrapAuthError(err)
	case http.StatusForbidden:
		return NewAppError(
			ErrorTypePermission,
			"Access forbidden: insufficient permissions",
			err,
		).WithSuggestion("Verify your permissions for this resource")
	case http.StatusNotFound:
		return NewAppError(
			ErrorTypeAPI,
			"Resource not found",
			err,
		).WithSuggestion("Verify the repository and issue numbers exist").
			WithSuggestion("Check if the repository is public or you have access")
	case http.StatusTooManyRequests:
		return NewAppError(
			ErrorTypeAPI,
			"API rate limit exceeded",
			err,
		).WithSuggestion("Wait a few minutes before retrying").
			WithSuggestion("Consider using authentication for higher rate limits")
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return NewAppError(
			ErrorTypeAPI,
			"GitHub API is temporarily unavailable",
			err,
		).WithSuggestion("Retry the operation in a few moments").
			WithSuggestion("Check GitHub's service status at https://www.githubstatus.com/")
	default:
		return NewAppError(
			ErrorTypeAPI,
			fmt.Sprintf("API request failed with status %d", statusCode),
			err,
		).WithContext("status_code", strconv.Itoa(statusCode))
	}
}

// Repository Errors
func NewRepositoryNotFoundError(repo string) *AppError {
	return NewAppError(
		ErrorTypeRepository,
		fmt.Sprintf("Repository not found: %s", repo),
		nil,
	).WithContext("repository", repo).
		WithSuggestion("Verify the repository name is spelled correctly").
		WithSuggestion("Check if the repository exists and is accessible to you").
		WithSuggestion("Use the --repo flag to specify a different repository")
}

func NewRepositoryAccessError(repo string, err error) *AppError {
	return NewAppError(
		ErrorTypeRepository,
		fmt.Sprintf("Cannot access repository: %s", repo),
		err,
	).WithContext("repository", repo).
		WithSuggestion("Verify you have access to this repository").
		WithSuggestion("Check if the repository is private and you're authenticated")
}

// Issue Errors
func NewIssueNotFoundError(repo string, issueNumber int) *AppError {
	return NewAppError(
		ErrorTypeIssue,
		fmt.Sprintf("Issue #%d not found in %s", issueNumber, repo),
		nil,
	).WithContext("repository", repo).
		WithContext("issue_number", strconv.Itoa(issueNumber)).
		WithSuggestion("Verify the issue number exists in the repository").
		WithSuggestion("Check if you have access to view the issue")
}

func NewDependencyExistsError(issueA, issueB string) *AppError {
	return NewAppError(
		ErrorTypeIssue,
		fmt.Sprintf("Dependency already exists between %s and %s", issueA, issueB),
		nil,
	).WithContext("issue_a", issueA).
		WithContext("issue_b", issueB).
		WithSuggestion("Use 'gh issue-dependency list' to view existing dependencies")
}

func NewCircularDependencyError(issueA, issueB string) *AppError {
	return NewAppError(
		ErrorTypeIssue,
		fmt.Sprintf("Cannot create circular dependency between %s and %s", issueA, issueB),
		nil,
	).WithContext("issue_a", issueA).
		WithContext("issue_b", issueB).
		WithSuggestion("Review the dependency chain to resolve circular references")
}

// Internal Errors
func WrapInternalError(operation string, err error) *AppError {
	return NewAppError(
		ErrorTypeInternal,
		fmt.Sprintf("Internal error during %s", operation),
		err,
	).WithContext("operation", operation).
		WithSuggestion("This appears to be a bug. Please report it with the error details")
}

// Error Formatting and User Display
func FormatUserError(err error) string {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		// Fallback for non-AppError types
		return fmt.Sprintf("Error: %v", err)
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Error: %s", appErr.Message))

	// Add context if available
	if len(appErr.Context) > 0 {
		output.WriteString("\n\nDetails:")
		for key, value := range appErr.Context {
			output.WriteString(fmt.Sprintf("\n  %s: %s", key, value))
		}
	}

	// Add suggestions if available
	if len(appErr.Suggestions) > 0 {
		output.WriteString("\n\nSuggestions:")
		for _, suggestion := range appErr.Suggestions {
			output.WriteString(fmt.Sprintf("\n  • %s", suggestion))
		}
	}

	return output.String()
}

// IsErrorType checks if an error is of a specific type
func IsErrorType(err error, errType ErrorType) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == errType
	}
	return false
}

// GetErrorType returns the error type, or ErrorTypeInternal if not an AppError
func GetErrorType(err error) ErrorType {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type
	}
	return ErrorTypeInternal
}

// Error handling utilities for common patterns

// HandleHTTPError converts HTTP response errors to appropriate AppErrors
func HandleHTTPError(resp *http.Response, operation string) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return WrapAuthError(fmt.Errorf("HTTP %d", resp.StatusCode))
	case http.StatusForbidden:
		return WrapPermissionError("repository", fmt.Errorf("HTTP %d", resp.StatusCode))
	case http.StatusNotFound:
		return NewAppError(
			ErrorTypeAPI,
			fmt.Sprintf("Resource not found during %s", operation),
			fmt.Errorf("HTTP %d", resp.StatusCode),
		).WithContext("operation", operation)
	default:
		return WrapAPIError(resp.StatusCode, fmt.Errorf("HTTP %d", resp.StatusCode))
	}
}

// ParseIssueReference parses various issue reference formats and validates them
func ParseIssueReference(ref string) (repo string, issueNum int, err error) {
	if ref == "" {
		return "", 0, NewEmptyValueError("issue reference")
	}

	// Handle numeric issue numbers
	if num, err := strconv.Atoi(ref); err == nil {
		if num <= 0 {
			return "", 0, NewIssueNumberValidationError(ref)
		}
		return "", num, nil
	}

	// Handle owner/repo#123 format
	if strings.Contains(ref, "#") {
		parts := strings.Split(ref, "#")
		if len(parts) != 2 {
			return "", 0, NewIssueNumberValidationError(ref)
		}

		repo = parts[0]
		if repo == "" {
			return "", 0, NewIssueNumberValidationError(ref)
		}

		num, err := strconv.Atoi(parts[1])
		if err != nil || num <= 0 {
			return "", 0, NewIssueNumberValidationError(ref)
		}

		// Validate repo format
		if !strings.Contains(repo, "/") {
			return "", 0, NewRepositoryFormatError(repo)
		}

		return repo, num, nil
	}

	// Handle GitHub URLs
	if strings.HasPrefix(ref, "https://github.com/") {
		// Extract from URL format: https://github.com/owner/repo/issues/123
		parts := strings.Split(strings.TrimPrefix(ref, "https://github.com/"), "/")
		if len(parts) < 4 || parts[2] != "issues" {
			return "", 0, NewIssueNumberValidationError(ref)
		}

		repo = parts[0] + "/" + parts[1]
		num, err := strconv.Atoi(parts[3])
		if err != nil || num <= 0 {
			return "", 0, NewIssueNumberValidationError(ref)
		}

		return repo, num, nil
	}

	return "", 0, NewIssueNumberValidationError(ref)
}

// ValidateRepository validates repository name format
func ValidateRepository(repo string) error {
	if repo == "" {
		return NewEmptyValueError("repository")
	}

	if !strings.Contains(repo, "/") {
		return NewRepositoryFormatError(repo)
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return NewRepositoryFormatError(repo)
	}

	return nil
}
