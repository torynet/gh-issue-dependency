// Package pkg provides validation utilities for dependency removal operations.
//
// This file implements comprehensive validation logic for removing GitHub issue
// dependency relationships, including relationship existence verification,
// permission checking, and input validation.
package pkg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
)

// RemovalValidator provides validation services for dependency removal operations.
// It leverages existing GitHub API utilities and error handling patterns to ensure
// safe and reliable dependency relationship removal.
type RemovalValidator struct {
	client *api.RESTClient
}

// NewRemovalValidator creates a new validator instance for dependency removal operations.
// It sets up the GitHub API client using the existing authentication patterns.
func NewRemovalValidator() (*RemovalValidator, error) {
	// Verify GitHub CLI authentication first
	if err := SetupGitHubClient(); err != nil {
		return nil, err
	}

	// Create GitHub API client using go-gh/v2 library
	client, err := api.DefaultRESTClient()
	if err != nil {
		return nil, WrapInternalError("creating GitHub API client for validation", err)
	}

	return &RemovalValidator{
		client: client,
	}, nil
}

// IssueRef represents a reference to a GitHub issue
type IssueRef struct {
	Owner  string
	Repo   string  
	Number int
	// FullName returns the full repository name (owner/repo)
	FullName string
}

// String returns a string representation of the issue reference
func (ref IssueRef) String() string {
	if ref.FullName != "" {
		return fmt.Sprintf("%s#%d", ref.FullName, ref.Number)
	}
	return fmt.Sprintf("%s/%s#%d", ref.Owner, ref.Repo, ref.Number)
}

// RemoveOptions contains options for dependency removal operations
type RemoveOptions struct {
	DryRun bool
	Force  bool
}

// ValidationResult contains the result of a validation operation
type ValidationResult struct {
	Valid       bool
	Error       error
	Issues      []ValidationIssue
	Suggestions []string
}

// ValidationIssue represents a specific validation problem
type ValidationIssue struct {
	Type        string
	Message     string
	IssueRef    IssueRef
	Suggestions []string
}

// ValidateRemoval performs comprehensive validation for dependency removal operations.
// This includes input validation, permission checking, issue accessibility validation,
// and relationship existence verification.
func (v *RemovalValidator) ValidateRemoval(source, target IssueRef, relType string) error {
	// 1. Validate basic inputs
	if err := v.validateInputs(source, target, relType); err != nil {
		return fmt.Errorf("input validation failed: %w", err)
	}

	// 2. Check permissions for source repository (where we'll be removing the relationship)
	if err := v.validatePermissions(source); err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}

	// 3. Verify issues exist and are accessible
	if err := v.validateIssueAccess(source, target); err != nil {
		return fmt.Errorf("issue access validation failed: %w", err)
	}

	// 4. Verify relationship actually exists
	exists, err := v.VerifyRelationshipExists(source, target, relType)
	if err != nil {
		return fmt.Errorf("relationship verification failed: %w", err)
	}
	if !exists {
		return v.createRelationshipNotFoundError(source, target, relType)
	}

	return nil
}

// ValidateBatchRemoval validates removal of multiple dependencies at once
func (v *RemovalValidator) ValidateBatchRemoval(source IssueRef, targets []IssueRef, relType string) error {
	// Check permissions once for the source repository
	if err := v.validatePermissions(source); err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}

	// Validate source issue accessibility
	if err := v.validateSingleIssueAccess(source); err != nil {
		return fmt.Errorf("source issue validation failed: %w", err)
	}

	// Validate each target and check relationship existence
	var validationErrors []string
	for _, target := range targets {
		// Validate target issue accessibility
		if err := v.validateSingleIssueAccess(target); err != nil {
			validationErrors = append(validationErrors, 
				fmt.Sprintf("Target issue %s: %v", target.String(), err))
			continue
		}

		// Verify relationship exists
		exists, err := v.VerifyRelationshipExists(source, target, relType)
		if err != nil {
			validationErrors = append(validationErrors, 
				fmt.Sprintf("Relationship verification failed for %s: %v", target.String(), err))
			continue
		}
		if !exists {
			validationErrors = append(validationErrors, 
				fmt.Sprintf("No %s relationship found between %s and %s", relType, source.String(), target.String()))
		}
	}

	if len(validationErrors) > 0 {
		return NewAppError(
			ErrorTypeValidation,
			"Batch validation failed",
			nil,
		).WithContext("errors", strings.Join(validationErrors, "; ")).
			WithSuggestion("Use 'gh issue-dependency list' to see current relationships").
			WithSuggestion("Verify issue numbers and repository access")
	}

	return nil
}

// VerifyRelationshipExists checks if a dependency relationship actually exists
// between the source and target issues using the GitHub API.
func (v *RemovalValidator) VerifyRelationshipExists(source, target IssueRef, relType string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch existing relationships for the source issue
	dependencies, err := v.fetchIssueDependencies(ctx, source)
	if err != nil {
		return false, fmt.Errorf("failed to fetch dependencies for %s: %w", source.String(), err)
	}

	// Check if the target exists in the appropriate relationship list
	return v.relationshipExistsInData(dependencies, target, relType), nil
}

// validateInputs performs basic input validation
func (v *RemovalValidator) validateInputs(source, target IssueRef, relType string) error {
	// Validate source issue reference
	if source.Owner == "" || source.Repo == "" || source.Number <= 0 {
		return NewEmptyValueError("source issue reference")
	}

	// Validate target issue reference  
	if target.Owner == "" || target.Repo == "" || target.Number <= 0 {
		return NewEmptyValueError("target issue reference")
	}

	// Validate relationship type
	if relType != "blocked-by" && relType != "blocks" {
		return NewAppError(
			ErrorTypeValidation,
			fmt.Sprintf("Invalid relationship type: %s", relType),
			nil,
		).WithContext("relationship_type", relType).
			WithSuggestion("Use either 'blocked-by' or 'blocks'")
	}

	// Prevent self-references
	if source.Owner == target.Owner && source.Repo == target.Repo && source.Number == target.Number {
		return NewAppError(
			ErrorTypeValidation,
			"Cannot remove dependency relationship from an issue to itself",
			nil,
		).WithContext("issue", source.String()).
			WithSuggestion("Specify different source and target issues")
	}

	return nil
}

// validatePermissions checks if the user has write permissions to modify relationships
func (v *RemovalValidator) validatePermissions(source IssueRef) error {
	// Use existing repository access validation from github.go
	repoName := fmt.Sprintf("%s/%s", source.Owner, source.Repo)
	if err := ValidateRepoAccess(source.Owner, source.Repo); err != nil {
		return err
	}

	// For dependency modification, we need write access to the source repository
	// The GitHub CLI approach checks general access, but for modifications we should
	// verify write permissions more specifically
	if err := v.validateWritePermissions(source); err != nil {
		return NewPermissionDeniedError("modify dependencies", repoName).
			WithSuggestion("Ensure you have write or maintain permissions for this repository").
			WithSuggestion("Contact the repository owner to request appropriate access")
	}

	return nil
}

// validateWritePermissions checks for write access to the repository
func (v *RemovalValidator) validateWritePermissions(ref IssueRef) error {
	// Check repository permissions via API
	endpoint := fmt.Sprintf("repos/%s/%s", ref.Owner, ref.Repo)
	
	var repoData struct {
		Permissions struct {
			Push  bool `json:"push"`
			Admin bool `json:"admin"`
		} `json:"permissions"`
	}

	err := v.client.Get(endpoint, &repoData)
	if err != nil {
		// If we can't check permissions, fall back to basic access validation
		return ValidateRepoAccess(ref.Owner, ref.Repo)
	}

	// Check if user has push or admin permissions (required for dependency modification)
	if !repoData.Permissions.Push && !repoData.Permissions.Admin {
		return NewPermissionDeniedError("modify dependencies", 
			fmt.Sprintf("%s/%s", ref.Owner, ref.Repo))
	}

	return nil
}

// validateIssueAccess validates that both source and target issues are accessible
func (v *RemovalValidator) validateIssueAccess(source, target IssueRef) error {
	// Validate source issue
	if err := v.validateSingleIssueAccess(source); err != nil {
		return fmt.Errorf("source issue %s: %w", source.String(), err)
	}

	// Validate target issue
	if err := v.validateSingleIssueAccess(target); err != nil {
		return fmt.Errorf("target issue %s: %w", target.String(), err)
	}

	return nil
}

// validateSingleIssueAccess validates access to a single issue
func (v *RemovalValidator) validateSingleIssueAccess(ref IssueRef) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use existing fetchIssueDetails function from github.go
	_, err := fetchIssueDetails(ctx, v.client, ref.Owner, ref.Repo, ref.Number)
	if err != nil {
		return err // fetchIssueDetails already returns appropriate error types
	}

	return nil
}

// fetchIssueDependencies fetches dependency data for an issue
func (v *RemovalValidator) fetchIssueDependencies(ctx context.Context, ref IssueRef) (*DependencyData, error) {
	// Use existing FetchIssueDependencies function from github.go
	return FetchIssueDependencies(ctx, ref.Owner, ref.Repo, ref.Number)
}

// relationshipExistsInData checks if a relationship exists in the dependency data
func (v *RemovalValidator) relationshipExistsInData(data *DependencyData, target IssueRef, relType string) bool {
	if data == nil {
		return false
	}

	var relationsToCheck []DependencyRelation
	
	switch relType {
	case "blocked-by":
		// For blocked-by relationship, check the BlockedBy list
		relationsToCheck = data.BlockedBy
	case "blocks": 
		// For blocks relationship, check the Blocking list
		relationsToCheck = data.Blocking
	}

	// Check if target exists in the relationship list
	for _, relation := range relationsToCheck {
		// Match by issue number and repository
		if relation.Issue.Number == target.Number {
			// Check repository match - handle both full name and individual owner/repo
			targetRepo := fmt.Sprintf("%s/%s", target.Owner, target.Repo)
			if target.FullName != "" {
				targetRepo = target.FullName
			}
			
			if relation.Repository == targetRepo || 
			   relation.Issue.Repository.FullName == targetRepo {
				return true
			}
		}
	}

	return false
}

// createRelationshipNotFoundError creates a specific error for non-existent relationships
func (v *RemovalValidator) createRelationshipNotFoundError(source, target IssueRef, relType string) error {
	var relationshipDescription string
	switch relType {
	case "blocked-by":
		relationshipDescription = fmt.Sprintf("%s is not blocked by %s", source.String(), target.String())
	case "blocks":
		relationshipDescription = fmt.Sprintf("%s does not block %s", source.String(), target.String())
	}

	return NewAppError(
		ErrorTypeIssue,
		fmt.Sprintf("Cannot remove dependency: %s", relationshipDescription),
		nil,
	).WithContext("source", source.String()).
		WithContext("target", target.String()).
		WithContext("relationship_type", relType).
		WithSuggestion(fmt.Sprintf("Use 'gh issue-dependency list %d' to see current dependencies", source.Number)).
		WithSuggestion("Verify the issue numbers and relationship type are correct")
}

// Helper functions for converting between issue reference formats

// ParseIssueRefWithRepo parses an issue reference string and creates an IssueRef
// It handles the same formats as the existing ParseIssueReference function
func ParseIssueRefWithRepo(issueRefStr, defaultOwner, defaultRepo string) (IssueRef, error) {
	if defaultOwner == "" || defaultRepo == "" {
		return IssueRef{}, NewEmptyValueError("default repository context")
	}

	// Use existing ParseIssueReference function
	repo, issueNum, err := ParseIssueReference(issueRefStr)
	if err != nil {
		return IssueRef{}, err
	}

	// If no repository specified, use default
	if repo == "" {
		return IssueRef{
			Owner:  defaultOwner,
			Repo:   defaultRepo, 
			Number: issueNum,
		}, nil
	}

	// Validate and parse repository
	if err := ValidateRepository(repo); err != nil {
		return IssueRef{}, err
	}

	parts := strings.Split(repo, "/")
	return IssueRef{
		Owner:    parts[0],
		Repo:     parts[1],
		Number:   issueNum,
		FullName: repo,
	}, nil
}

// CreateIssueRef creates an IssueRef from individual components
func CreateIssueRef(owner, repo string, number int) IssueRef {
	return IssueRef{
		Owner:    owner,
		Repo:     repo,
		Number:   number,
		FullName: fmt.Sprintf("%s/%s", owner, repo),
	}
}