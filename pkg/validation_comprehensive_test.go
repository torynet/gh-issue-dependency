package pkg

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateInputs tests the input validation function with comprehensive edge cases
func TestValidateInputs(t *testing.T) {
	tests := []struct {
		name           string
		source         IssueRef
		target         IssueRef
		relType        string
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "valid blocked-by relationship",
			source: IssueRef{
				Owner:  "owner",
				Repo:   "repo",
				Number: 123,
			},
			target: IssueRef{
				Owner:  "owner",
				Repo:   "repo",
				Number: 456,
			},
			relType:     "blocked-by",
			expectError: false,
		},
		{
			name: "valid blocks relationship",
			source: IssueRef{
				Owner:  "owner",
				Repo:   "repo",
				Number: 123,
			},
			target: IssueRef{
				Owner:  "other",
				Repo:   "repo",
				Number: 456,
			},
			relType:     "blocks",
			expectError: false,
		},
		{
			name: "empty source owner",
			source: IssueRef{
				Owner:  "",
				Repo:   "repo",
				Number: 123,
			},
			target: IssueRef{
				Owner:  "owner",
				Repo:   "repo",
				Number: 456,
			},
			relType:        "blocked-by",
			expectError:    true,
			expectedErrMsg: "source issue reference",
		},
		{
			name: "empty source repo",
			source: IssueRef{
				Owner:  "owner",
				Repo:   "",
				Number: 123,
			},
			target: IssueRef{
				Owner:  "owner",
				Repo:   "repo",
				Number: 456,
			},
			relType:        "blocked-by",
			expectError:    true,
			expectedErrMsg: "source issue reference",
		},
		{
			name: "invalid source issue number - zero",
			source: IssueRef{
				Owner:  "owner",
				Repo:   "repo",
				Number: 0,
			},
			target: IssueRef{
				Owner:  "owner",
				Repo:   "repo",
				Number: 456,
			},
			relType:        "blocked-by",
			expectError:    true,
			expectedErrMsg: "source issue reference",
		},
		{
			name: "invalid source issue number - negative",
			source: IssueRef{
				Owner:  "owner",
				Repo:   "repo",
				Number: -1,
			},
			target: IssueRef{
				Owner:  "owner",
				Repo:   "repo",
				Number: 456,
			},
			relType:        "blocked-by",
			expectError:    true,
			expectedErrMsg: "source issue reference",
		},
		{
			name: "empty target owner",
			source: IssueRef{
				Owner:  "owner",
				Repo:   "repo",
				Number: 123,
			},
			target: IssueRef{
				Owner:  "",
				Repo:   "repo",
				Number: 456,
			},
			relType:        "blocked-by",
			expectError:    true,
			expectedErrMsg: "target issue reference",
		},
		{
			name: "empty target repo",
			source: IssueRef{
				Owner:  "owner",
				Repo:   "repo",
				Number: 123,
			},
			target: IssueRef{
				Owner:  "owner",
				Repo:   "",
				Number: 456,
			},
			relType:        "blocked-by",
			expectError:    true,
			expectedErrMsg: "target issue reference",
		},
		{
			name: "invalid target issue number",
			source: IssueRef{
				Owner:  "owner",
				Repo:   "repo",
				Number: 123,
			},
			target: IssueRef{
				Owner:  "owner",
				Repo:   "repo",
				Number: 0,
			},
			relType:        "blocked-by",
			expectError:    true,
			expectedErrMsg: "target issue reference",
		},
		{
			name: "invalid relationship type",
			source: IssueRef{
				Owner:  "owner",
				Repo:   "repo",
				Number: 123,
			},
			target: IssueRef{
				Owner:  "owner",
				Repo:   "repo",
				Number: 456,
			},
			relType:        "depends-on",
			expectError:    true,
			expectedErrMsg: "Invalid relationship type",
		},
		{
			name: "self-reference blocked-by",
			source: IssueRef{
				Owner:  "owner",
				Repo:   "repo",
				Number: 123,
			},
			target: IssueRef{
				Owner:  "owner",
				Repo:   "repo",
				Number: 123,
			},
			relType:        "blocked-by",
			expectError:    true,
			expectedErrMsg: "Cannot remove dependency relationship from an issue to itself",
		},
		{
			name: "self-reference blocks",
			source: IssueRef{
				Owner:  "owner",
				Repo:   "repo",
				Number: 456,
			},
			target: IssueRef{
				Owner:  "owner",
				Repo:   "repo",
				Number: 456,
			},
			relType:        "blocks",
			expectError:    true,
			expectedErrMsg: "Cannot remove dependency relationship from an issue to itself",
		},
	}

	// Create a validator instance for testing (without API client)
	v := &RemovalValidator{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.validateInputs(tt.source, tt.target, tt.relType)

			if tt.expectError {
				require.Error(t, err, "Expected error but got none")
				assert.Contains(t, err.Error(), tt.expectedErrMsg,
					"Error message should contain: %s", tt.expectedErrMsg)
			} else {
				assert.NoError(t, err, "Unexpected error: %v", err)
			}
		})
	}
}

// TestParseIssueRefWithRepo tests issue reference parsing with comprehensive cases
func TestParseIssueRefWithRepo(t *testing.T) {
	tests := []struct {
		name         string
		issueRef     string
		defaultOwner string
		defaultRepo  string
		expected     IssueRef
		expectError  bool
	}{
		{
			name:         "Simple issue number",
			issueRef:     "123",
			defaultOwner: "testowner",
			defaultRepo:  "testrepo",
			expected: IssueRef{
				Owner:  "testowner",
				Repo:   "testrepo",
				Number: 123,
			},
			expectError: false,
		},
		{
			name:         "Cross-repo reference",
			issueRef:     "otherowner/otherrepo#456",
			defaultOwner: "testowner",
			defaultRepo:  "testrepo",
			expected: IssueRef{
				Owner:    "otherowner",
				Repo:     "otherrepo",
				Number:   456,
				FullName: "otherowner/otherrepo",
			},
			expectError: false,
		},
		{
			name:         "GitHub URL",
			issueRef:     "https://github.com/testowner/testrepo/issues/789",
			defaultOwner: "defaultowner",
			defaultRepo:  "defaultrepo",
			expected: IssueRef{
				Owner:    "testowner",
				Repo:     "testrepo",
				Number:   789,
				FullName: "testowner/testrepo",
			},
			expectError: false,
		},
		{
			name:         "Invalid issue number",
			issueRef:     "not-a-number",
			defaultOwner: "testowner",
			defaultRepo:  "testrepo",
			expected:     IssueRef{},
			expectError:  true,
		},
		{
			name:         "Empty default owner",
			issueRef:     "123",
			defaultOwner: "",
			defaultRepo:  "testrepo",
			expected:     IssueRef{},
			expectError:  true,
		},
		{
			name:         "Empty default repo",
			issueRef:     "123",
			defaultOwner: "testowner",
			defaultRepo:  "",
			expected:     IssueRef{},
			expectError:  true,
		},
		{
			name:         "Zero issue number",
			issueRef:     "0",
			defaultOwner: "testowner",
			defaultRepo:  "testrepo",
			expected:     IssueRef{},
			expectError:  true,
		},
		{
			name:         "Negative issue number",
			issueRef:     "-123",
			defaultOwner: "testowner",
			defaultRepo:  "testrepo",
			expected:     IssueRef{},
			expectError:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseIssueRefWithRepo(tt.issueRef, tt.defaultOwner, tt.defaultRepo)
			
			if tt.expectError {
				require.Error(t, err, "Expected error but got none")
			} else {
				require.NoError(t, err, "Unexpected error: %v", err)
				assert.Equal(t, tt.expected.Owner, result.Owner, "Owner should match")
				assert.Equal(t, tt.expected.Repo, result.Repo, "Repo should match")
				assert.Equal(t, tt.expected.Number, result.Number, "Number should match")
				assert.Equal(t, tt.expected.FullName, result.FullName, "FullName should match")
			}
		})
	}
}

// TestCreateIssueRef tests the CreateIssueRef helper function
func TestCreateIssueRef(t *testing.T) {
	tests := []struct {
		name     string
		owner    string
		repo     string
		number   int
		expected IssueRef
	}{
		{
			name:   "valid issue reference",
			owner:  "owner",
			repo:   "repo",
			number: 123,
			expected: IssueRef{
				Owner:    "owner",
				Repo:     "repo",
				Number:   123,
				FullName: "owner/repo",
			},
		},
		{
			name:   "complex repository names",
			owner:  "complex-owner-name",
			repo:   "complex.repo-name_test",
			number: 9999,
			expected: IssueRef{
				Owner:    "complex-owner-name",
				Repo:     "complex.repo-name_test",
				Number:   9999,
				FullName: "complex-owner-name/complex.repo-name_test",
			},
		},
		{
			name:   "single character names",
			owner:  "a",
			repo:   "b",
			number: 1,
			expected: IssueRef{
				Owner:    "a",
				Repo:     "b",
				Number:   1,
				FullName: "a/b",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateIssueRef(tt.owner, tt.repo, tt.number)
			assert.Equal(t, tt.expected, result, "CreateIssueRef result should match expected")
		})
	}
}

// TestIssueRefString tests the String method of IssueRef
func TestIssueRefString(t *testing.T) {
	tests := []struct {
		name     string
		ref      IssueRef
		expected string
	}{
		{
			name: "With FullName",
			ref: IssueRef{
				Owner:    "owner",
				Repo:     "repo",
				Number:   123,
				FullName: "owner/repo",
			},
			expected: "owner/repo#123",
		},
		{
			name: "Without FullName",
			ref: IssueRef{
				Owner:  "owner",
				Repo:   "repo",
				Number: 123,
			},
			expected: "owner/repo#123",
		},
		{
			name: "Complex repository names",
			ref: IssueRef{
				Owner:    "complex-owner",
				Repo:     "complex.repo-name",
				Number:   9999,
				FullName: "complex-owner/complex.repo-name",
			},
			expected: "complex-owner/complex.repo-name#9999",
		},
		{
			name: "Zero issue number",
			ref: IssueRef{
				Owner:  "owner",
				Repo:   "repo",
				Number: 0,
			},
			expected: "owner/repo#0",
		},
		{
			name: "Large issue number",
			ref: IssueRef{
				Owner:    "owner",
				Repo:     "repo",
				Number:   999999,
				FullName: "owner/repo",
			},
			expected: "owner/repo#999999",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ref.String()
			assert.Equal(t, tt.expected, result, "String() result should match expected")
		})
	}
}

// TestRelationshipExistsInData tests the relationship existence checking logic
func TestRelationshipExistsInData(t *testing.T) {
	// Create a validator instance for testing
	v := &RemovalValidator{}

	// Create test dependency data
	testData := &DependencyData{
		SourceIssue: Issue{
			Number:     123,
			Title:      "Source Issue",
			State:      "open",
			Repository: RepositoryInfo{FullName: "owner/repo"},
		},
		BlockedBy: []DependencyRelation{
			{
				Issue: Issue{
					Number:     456,
					Title:      "Blocker Issue 1",
					State:      "open",
					Repository: RepositoryInfo{FullName: "owner/repo"},
				},
				Type:       "blocked_by",
				Repository: "owner/repo",
			},
			{
				Issue: Issue{
					Number:     789,
					Title:      "Cross-repo Blocker",
					State:      "open",
					Repository: RepositoryInfo{FullName: "other/repo"},
				},
				Type:       "blocked_by",
				Repository: "other/repo",
			},
		},
		Blocking: []DependencyRelation{
			{
				Issue: Issue{
					Number:     101,
					Title:      "Blocked Issue 1",
					State:      "open",
					Repository: RepositoryInfo{FullName: "owner/repo"},
				},
				Type:       "blocks",
				Repository: "owner/repo",
			},
		},
		FetchedAt:  time.Now(),
		TotalCount: 3,
	}

	tests := []struct {
		name       string
		data       *DependencyData
		target     IssueRef
		relType    string
		expected   bool
	}{
		{
			name:   "existing blocked-by relationship - same repo",
			data:   testData,
			target: CreateIssueRef("owner", "repo", 456),
			relType: "blocked-by",
			expected: true,
		},
		{
			name:   "existing blocked-by relationship - cross repo",
			data:   testData,
			target: CreateIssueRef("other", "repo", 789),
			relType: "blocked-by",
			expected: true,
		},
		{
			name:   "existing blocks relationship",
			data:   testData,
			target: CreateIssueRef("owner", "repo", 101),
			relType: "blocks",
			expected: true,
		},
		{
			name:   "non-existing blocked-by relationship",
			data:   testData,
			target: CreateIssueRef("owner", "repo", 999),
			relType: "blocked-by",
			expected: false,
		},
		{
			name:   "non-existing blocks relationship",
			data:   testData,
			target: CreateIssueRef("owner", "repo", 999),
			relType: "blocks",
			expected: false,
		},
		{
			name:   "wrong relationship type - blocks for blocked-by target",
			data:   testData,
			target: CreateIssueRef("owner", "repo", 456),
			relType: "blocks",
			expected: false,
		},
		{
			name:   "wrong relationship type - blocked-by for blocks target",
			data:   testData,
			target: CreateIssueRef("owner", "repo", 101),
			relType: "blocked-by",
			expected: false,
		},
		{
			name:   "nil data",
			data:   nil,
			target: CreateIssueRef("owner", "repo", 456),
			relType: "blocked-by",
			expected: false,
		},
		{
			name: "empty data",
			data: &DependencyData{
				SourceIssue: Issue{Number: 123},
				BlockedBy:   []DependencyRelation{},
				Blocking:    []DependencyRelation{},
			},
			target:   CreateIssueRef("owner", "repo", 456),
			relType:  "blocked-by",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.relationshipExistsInData(tt.data, tt.target, tt.relType)
			assert.Equal(t, tt.expected, result,
				"Relationship existence check should return %v for %s", tt.expected, tt.name)
		})
	}
}

// TestCreateRelationshipNotFoundError tests error creation for non-existent relationships
func TestCreateRelationshipNotFoundError(t *testing.T) {
	v := &RemovalValidator{}

	tests := []struct {
		name              string
		source            IssueRef
		target            IssueRef
		relType           string
		expectedErrMsg    string
		expectedErrType   ErrorType
	}{
		{
			name:            "blocked-by relationship not found",
			source:          CreateIssueRef("owner", "repo", 123),
			target:          CreateIssueRef("owner", "repo", 456),
			relType:         "blocked-by",
			expectedErrMsg:  "owner/repo#123 is not blocked by owner/repo#456",
			expectedErrType: ErrorTypeIssue,
		},
		{
			name:            "blocks relationship not found",
			source:          CreateIssueRef("owner", "repo", 123),
			target:          CreateIssueRef("other", "repo", 789),
			relType:         "blocks",
			expectedErrMsg:  "owner/repo#123 does not block other/repo#789",
			expectedErrType: ErrorTypeIssue,
		},
		{
			name:            "cross-repository blocked-by not found",
			source:          CreateIssueRef("source", "repo", 100),
			target:          CreateIssueRef("target", "repo", 200),
			relType:         "blocked-by",
			expectedErrMsg:  "source/repo#100 is not blocked by target/repo#200",
			expectedErrType: ErrorTypeIssue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.createRelationshipNotFoundError(tt.source, tt.target, tt.relType)

			require.Error(t, err, "Should return an error")
			assert.Contains(t, err.Error(), tt.expectedErrMsg,
				"Error message should contain expected text")
			
			// Check if it's an AppError and has the correct type
			if appErr, ok := err.(*AppError); ok {
				assert.Equal(t, tt.expectedErrType, appErr.Type,
					"Error type should match expected")
			}
		})
	}
}

// TestRemoveOptions tests the RemoveOptions struct
func TestRemoveOptions(t *testing.T) {
	tests := []struct {
		name     string
		opts     RemoveOptions
		expected RemoveOptions
	}{
		{
			name: "default options",
			opts: RemoveOptions{},
			expected: RemoveOptions{
				DryRun: false,
				Force:  false,
			},
		},
		{
			name: "dry run mode",
			opts: RemoveOptions{
				DryRun: true,
				Force:  false,
			},
			expected: RemoveOptions{
				DryRun: true,
				Force:  false,
			},
		},
		{
			name: "force mode",
			opts: RemoveOptions{
				DryRun: false,
				Force:  true,
			},
			expected: RemoveOptions{
				DryRun: false,
				Force:  true,
			},
		},
		{
			name: "both flags (edge case)",
			opts: RemoveOptions{
				DryRun: true,
				Force:  true,
			},
			expected: RemoveOptions{
				DryRun: true,
				Force:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected.DryRun, tt.opts.DryRun, "DryRun should match")
			assert.Equal(t, tt.expected.Force, tt.opts.Force, "Force should match")
		})
	}
}

// TestValidationResult tests the ValidationResult struct
func TestValidationResult(t *testing.T) {
	tests := []struct {
		name     string
		result   ValidationResult
		expected bool
	}{
		{
			name: "valid result",
			result: ValidationResult{
				Valid: true,
				Error: nil,
				Issues: []ValidationIssue{},
				Suggestions: []string{},
			},
			expected: true,
		},
		{
			name: "invalid result with error",
			result: ValidationResult{
				Valid: false,
				Error: fmt.Errorf("validation error"),
				Issues: []ValidationIssue{
					{
						Type:    "error",
						Message: "Test error",
						IssueRef: CreateIssueRef("owner", "repo", 123),
						Suggestions: []string{"Fix the error"},
					},
				},
				Suggestions: []string{"General suggestion"},
			},
			expected: false,
		},
		{
			name: "valid result with suggestions",
			result: ValidationResult{
				Valid: true,
				Error: nil,
				Issues: []ValidationIssue{},
				Suggestions: []string{"Consider this improvement"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.Valid, "Valid field should match")
			
			if !tt.expected {
				assert.Error(t, tt.result.Error, "Should have an error when invalid")
				assert.NotEmpty(t, tt.result.Issues, "Should have validation issues when invalid")
			}
			
			if len(tt.result.Suggestions) > 0 {
				assert.NotEmpty(t, tt.result.Suggestions[0], "Suggestions should not be empty")
			}
		})
	}
}

// TestValidationIssue tests the ValidationIssue struct
func TestValidationIssue(t *testing.T) {
	tests := []struct {
		name  string
		issue ValidationIssue
	}{
		{
			name: "complete validation issue",
			issue: ValidationIssue{
				Type:        "permission",
				Message:     "Insufficient permissions",
				IssueRef:    CreateIssueRef("owner", "repo", 123),
				Suggestions: []string{"Request write access", "Contact repository owner"},
			},
		},
		{
			name: "minimal validation issue",
			issue: ValidationIssue{
				Type:     "input",
				Message:  "Invalid input",
				IssueRef: IssueRef{},
			},
		},
		{
			name: "validation issue with cross-repo reference",
			issue: ValidationIssue{
				Type:        "access",
				Message:     "Cannot access cross-repository issue",
				IssueRef:    CreateIssueRef("other", "repo", 456),
				Suggestions: []string{"Check repository visibility", "Verify issue exists"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.issue.Type, "Type should not be empty")
			assert.NotEmpty(t, tt.issue.Message, "Message should not be empty")
			
			// If IssueRef is set, validate its string representation
			if tt.issue.IssueRef.Number > 0 {
				issueStr := tt.issue.IssueRef.String()
				assert.Contains(t, issueStr, "#", "Issue string should contain #")
			}
			
			// If suggestions exist, validate they're not empty
			for i, suggestion := range tt.issue.Suggestions {
				assert.NotEmpty(t, suggestion, "Suggestion %d should not be empty", i)
			}
		})
	}
}

// TestValidationErrorScenarios demonstrates comprehensive error handling patterns
func TestValidationErrorScenarios(t *testing.T) {
	t.Run("Input validation error scenarios", func(t *testing.T) {
		scenarios := []struct {
			name        string
			description string
			errorType   ErrorType
			setupError  func() error
		}{
			{
				name:        "Empty source issue",
				description: "Should return validation error for empty source",
				errorType:   ErrorTypeValidation,
				setupError: func() error {
					v := &RemovalValidator{}
					emptySource := IssueRef{}
					validTarget := CreateIssueRef("owner", "repo", 456)
					return v.validateInputs(emptySource, validTarget, "blocked-by")
				},
			},
			{
				name:        "Invalid relationship type",
				description: "Should return validation error for invalid relationship type",
				errorType:   ErrorTypeValidation,
				setupError: func() error {
					v := &RemovalValidator{}
					source := CreateIssueRef("owner", "repo", 123)
					target := CreateIssueRef("owner", "repo", 456)
					return v.validateInputs(source, target, "invalid-type")
				},
			},
			{
				name:        "Self-reference",
				description: "Should return validation error for self-reference",
				errorType:   ErrorTypeValidation,
				setupError: func() error {
					v := &RemovalValidator{}
					selfRef := CreateIssueRef("owner", "repo", 123)
					return v.validateInputs(selfRef, selfRef, "blocked-by")
				},
			},
			{
				name:        "Relationship not found",
				description: "Should return issue error for non-existent relationship",
				errorType:   ErrorTypeIssue,
				setupError: func() error {
					v := &RemovalValidator{}
					source := CreateIssueRef("owner", "repo", 123)
					target := CreateIssueRef("owner", "repo", 456)
					return v.createRelationshipNotFoundError(source, target, "blocked-by")
				},
			},
		}
		
		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				err := scenario.setupError()
				
				require.Error(t, err, "Scenario should produce an error")
				
				// Check error type if it's an AppError
				if appErr, ok := err.(*AppError); ok {
					assert.Equal(t, scenario.errorType, appErr.Type,
						"Error type should match expected for: %s", scenario.name)
				}
				
				t.Logf("Scenario: %s - %s", scenario.name, scenario.description)
				t.Logf("Expected error type: %v", scenario.errorType)
				t.Logf("Actual error: %v", err)
			})
		}
	})
}

// Example usage patterns for integration with cmd/remove.go
func ExampleRemovalValidator_ValidateRemoval() {
	// This example shows how to use the validator in the remove command
	
	// 1. Create validator (would require GitHub CLI auth in real usage)
	// validator, err := NewRemovalValidator()
	// if err != nil {
	// 	fmt.Printf("Failed to create validator: %v\n", err)
	// 	return
	// }
	
	// 2. Parse issue references
	source := CreateIssueRef("owner", "repo", 123)
	target := CreateIssueRef("owner", "repo", 456)
	
	// 3. Validate removal (would require API calls in real usage)
	// err = validator.ValidateRemoval(source, target, "blocked-by")
	// if err != nil {
	// 	fmt.Printf("Validation failed: %v\n", err)
	// 	return
	// }
	
	fmt.Printf("Validation example: %s -> %s (blocked-by)\n", source.String(), target.String())
	// Output: Validation example: owner/repo#123 -> owner/repo#456 (blocked-by)
}

// Example batch validation
func ExampleRemovalValidator_ValidateBatchRemoval() {
	// validator, err := NewRemovalValidator()
	// if err != nil {
	// 	fmt.Printf("Failed to create validator: %v\n", err)
	// 	return
	// }
	
	source := CreateIssueRef("owner", "repo", 123)
	targets := []IssueRef{
		CreateIssueRef("owner", "repo", 456),
		CreateIssueRef("owner", "repo", 789),
		CreateIssueRef("other", "repo", 101),
	}
	
	// err = validator.ValidateBatchRemoval(source, targets, "blocks")
	// if err != nil {
	// 	fmt.Printf("Batch validation failed: %v\n", err)
	// 	return
	// }
	
	fmt.Printf("Batch validation example: %s blocks %d issues\n", source.String(), len(targets))
	// Output: Batch validation example: owner/repo#123 blocks 3 issues
}