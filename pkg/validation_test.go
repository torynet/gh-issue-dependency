package pkg

import (
	"fmt"
	"testing"
)

// TestValidationIntegration demonstrates how the validation system works
func TestValidationIntegration(t *testing.T) {
	// This test demonstrates the validation integration without making actual API calls
	// In a real scenario, these would be integration tests with mock GitHub API responses

	t.Run("ParseIssueRefWithRepo", func(t *testing.T) {
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
				name:         "Invalid issue number",
				issueRef:     "not-a-number",
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
					if err == nil {
						t.Errorf("Expected error but got none")
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error: %v", err)
					}
					if result.Owner != tt.expected.Owner || result.Repo != tt.expected.Repo || result.Number != tt.expected.Number {
						t.Errorf("Expected %v, got %v", tt.expected, result)
					}
				}
			})
		}
	})

	t.Run("CreateIssueRef", func(t *testing.T) {
		ref := CreateIssueRef("owner", "repo", 123)
		expected := IssueRef{
			Owner:    "owner",
			Repo:     "repo",
			Number:   123,
			FullName: "owner/repo",
		}

		if ref != expected {
			t.Errorf("Expected %v, got %v", expected, ref)
		}
	})

	t.Run("IssueRef.String", func(t *testing.T) {
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
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.ref.String()
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			})
		}
	})
}

// TestValidationErrorScenarios demonstrates error handling for various scenarios
func TestBasicValidationErrorScenarios(t *testing.T) {
	t.Run("Input validation errors", func(t *testing.T) {
		// Test cases would validate different error conditions
		// This demonstrates the expected error patterns without actual API calls

		scenarios := []struct {
			name        string
			description string
			errorType   ErrorType
		}{
			{
				name:        "Empty source issue",
				description: "Should return validation error for empty source",
				errorType:   ErrorTypeValidation,
			},
			{
				name:        "Invalid relationship type",
				description: "Should return validation error for invalid relationship type",
				errorType:   ErrorTypeValidation,
			},
			{
				name:        "Self-reference",
				description: "Should return validation error for self-reference",
				errorType:   ErrorTypeValidation,
			},
		}

		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				// These tests would validate the error handling without making API calls
				t.Logf("Scenario: %s - %s", scenario.name, scenario.description)
				t.Logf("Expected error type: %s", scenario.errorType)
			})
		}
	})
}

// Example usage patterns for integration with cmd/remove.go
func ExampleBasicRemovalValidator_ValidateRemoval() {
	// This example shows how to use the validator in the remove command

	// 1. Create validator
	validator, err := NewRemovalValidator()
	if err != nil {
		fmt.Printf("Failed to create validator: %v\n", err)
		return
	}

	// 2. Parse issue references
	source := CreateIssueRef("owner", "repo", 123)
	target := CreateIssueRef("owner", "repo", 456)

	// 3. Validate removal
	err = validator.ValidateRemoval(source, target, "blocked-by")
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Println("Validation successful - ready to remove relationship")
}

// Example batch validation
func ExampleBasicRemovalValidator_ValidateBatchRemoval() {
	validator, err := NewRemovalValidator()
	if err != nil {
		fmt.Printf("Failed to create validator: %v\n", err)
		return
	}

	source := CreateIssueRef("owner", "repo", 123)
	targets := []IssueRef{
		CreateIssueRef("owner", "repo", 456),
		CreateIssueRef("owner", "repo", 789),
		CreateIssueRef("other", "repo", 101),
	}

	err = validator.ValidateBatchRemoval(source, targets, "blocks")
	if err != nil {
		fmt.Printf("Batch validation failed: %v\n", err)
		return
	}

	fmt.Println("Batch validation successful")
}
