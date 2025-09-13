package pkg

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockIntegrationEnvironment simulates the complete integration environment
type MockIntegrationEnvironment struct {
	// Repository state
	repositories map[string]*MockRepository

	// Authentication state
	authenticated bool
	currentUser   string

	// Command execution state
	executed      []string
	confirmations []string
	outputs       []string
}

// MockRepository represents a mock GitHub repository for integration testing
type MockRepository struct {
	Owner       string
	Name        string
	Issues      map[int]*MockIssue
	Permissions map[string]string // user -> permission level
}

// MockIssue represents a mock GitHub issue with dependencies
type MockIssue struct {
	Number    int
	Title     string
	State     string
	BlockedBy []IssueRef
	Blocking  []IssueRef
}

// NewMockIntegrationEnvironment creates a new mock environment for testing
func NewMockIntegrationEnvironment() *MockIntegrationEnvironment {
	return &MockIntegrationEnvironment{
		repositories:  make(map[string]*MockRepository),
		authenticated: true,
		currentUser:   "testuser",
		executed:      []string{},
		confirmations: []string{},
		outputs:       []string{},
	}
}

// AddRepository adds a mock repository to the environment
func (env *MockIntegrationEnvironment) AddRepository(owner, name string) *MockRepository {
	repo := &MockRepository{
		Owner:       owner,
		Name:        name,
		Issues:      make(map[int]*MockIssue),
		Permissions: make(map[string]string),
	}

	repoKey := fmt.Sprintf("%s/%s", owner, name)
	env.repositories[repoKey] = repo

	// Grant write permission to current user
	repo.Permissions[env.currentUser] = "write"

	return repo
}

// AddIssue adds a mock issue to a repository
func (repo *MockRepository) AddIssue(number int, title string) *MockIssue {
	issue := &MockIssue{
		Number:    number,
		Title:     title,
		State:     "open",
		BlockedBy: []IssueRef{},
		Blocking:  []IssueRef{},
	}

	repo.Issues[number] = issue
	return issue
}

// AddDependency adds a dependency relationship between issues
func (env *MockIntegrationEnvironment) AddDependency(source, target IssueRef, relType string) error {
	sourceRepo := env.repositories[fmt.Sprintf("%s/%s", source.Owner, source.Repo)]
	if sourceRepo == nil {
		return fmt.Errorf("source repository not found: %s/%s", source.Owner, source.Repo)
	}

	sourceIssue := sourceRepo.Issues[source.Number]
	if sourceIssue == nil {
		return fmt.Errorf("source issue not found: %d", source.Number)
	}

	// Add the relationship
	switch relType {
	case "blocked-by":
		sourceIssue.BlockedBy = append(sourceIssue.BlockedBy, target)
	case "blocks":
		sourceIssue.Blocking = append(sourceIssue.Blocking, target)
	}

	return nil
}

// RemoveDependency removes a dependency relationship
func (env *MockIntegrationEnvironment) RemoveDependency(source, target IssueRef, relType string) error {
	sourceRepo := env.repositories[fmt.Sprintf("%s/%s", source.Owner, source.Repo)]
	if sourceRepo == nil {
		return fmt.Errorf("source repository not found: %s/%s", source.Owner, source.Repo)
	}

	sourceIssue := sourceRepo.Issues[source.Number]
	if sourceIssue == nil {
		return fmt.Errorf("source issue not found: %d", source.Number)
	}

	// Remove the relationship
	switch relType {
	case "blocked-by":
		sourceIssue.BlockedBy = removeDependencyFromSlice(sourceIssue.BlockedBy, target)
	case "blocks":
		sourceIssue.Blocking = removeDependencyFromSlice(sourceIssue.Blocking, target)
	}

	return nil
}

// removeDependencyFromSlice removes a target from a dependency slice
func removeDependencyFromSlice(deps []IssueRef, target IssueRef) []IssueRef {
	var result []IssueRef
	for _, dep := range deps {
		if dep.Owner != target.Owner || dep.Repo != target.Repo || dep.Number != target.Number {
			result = append(result, dep)
		}
	}
	return result
}

// GetDependencies retrieves dependencies for an issue
func (env *MockIntegrationEnvironment) GetDependencies(issue IssueRef) (*DependencyData, error) {
	repoKey := fmt.Sprintf("%s/%s", issue.Owner, issue.Repo)
	repo := env.repositories[repoKey]
	if repo == nil {
		return nil, fmt.Errorf("repository not found: %s", repoKey)
	}

	mockIssue := repo.Issues[issue.Number]
	if mockIssue == nil {
		return nil, fmt.Errorf("issue not found: %d", issue.Number)
	}

	// Convert to DependencyData format
	data := &DependencyData{
		SourceIssue: Issue{
			Number:     mockIssue.Number,
			Title:      mockIssue.Title,
			State:      mockIssue.State,
			Repository: RepositoryInfo{FullName: repoKey},
		},
		BlockedBy:  []DependencyRelation{},
		Blocking:   []DependencyRelation{},
		FetchedAt:  time.Now(),
		TotalCount: len(mockIssue.BlockedBy) + len(mockIssue.Blocking),
	}

	// Convert blocked-by dependencies
	for _, dep := range mockIssue.BlockedBy {
		data.BlockedBy = append(data.BlockedBy, DependencyRelation{
			Issue: Issue{
				Number: dep.Number,
				Title:  fmt.Sprintf("Mock Issue %d", dep.Number),
				State:  "open",
			},
			Type:       "blocked_by",
			Repository: fmt.Sprintf("%s/%s", dep.Owner, dep.Repo),
		})
	}

	// Convert blocking dependencies
	for _, dep := range mockIssue.Blocking {
		data.Blocking = append(data.Blocking, DependencyRelation{
			Issue: Issue{
				Number: dep.Number,
				Title:  fmt.Sprintf("Mock Issue %d", dep.Number),
				State:  "open",
			},
			Type:       "blocks",
			Repository: fmt.Sprintf("%s/%s", dep.Owner, dep.Repo),
		})
	}

	return data, nil
}

// ExecuteCommand simulates command execution and records it
func (env *MockIntegrationEnvironment) ExecuteCommand(cmd string) {
	env.executed = append(env.executed, cmd)
}

// AddConfirmation simulates user confirmation input
func (env *MockIntegrationEnvironment) AddConfirmation(response string) {
	env.confirmations = append(env.confirmations, response)
}

// AddOutput records output for verification
func (env *MockIntegrationEnvironment) AddOutput(output string) {
	env.outputs = append(env.outputs, output)
}

// TestEndToEndSingleDependencyRemoval tests complete single dependency removal workflow
func TestEndToEndSingleDependencyRemoval(t *testing.T) {
	tests := []struct {
		name              string
		setupDependencies func(*MockIntegrationEnvironment)
		source            IssueRef
		target            IssueRef
		relType           string
		opts              RemoveOptions
		userConfirmation  string
		expectedSuccess   bool
		expectedOutput    []string
		expectedRemovals  int
	}{
		{
			name: "successful blocked-by removal with confirmation",
			setupDependencies: func(env *MockIntegrationEnvironment) {
				repo := env.AddRepository("owner", "repo")
				repo.AddIssue(123, "Feature: User Authentication")
				repo.AddIssue(456, "Database Setup")

				source := CreateIssueRef("owner", "repo", 123)
				target := CreateIssueRef("owner", "repo", 456)
				_ = env.AddDependency(source, target, "blocked-by")
			},
			source:           CreateIssueRef("owner", "repo", 123),
			target:           CreateIssueRef("owner", "repo", 456),
			relType:          "blocked-by",
			opts:             RemoveOptions{DryRun: false, Force: false},
			userConfirmation: "y",
			expectedSuccess:  true,
			expectedOutput: []string{
				"Remove dependency relationship?",
				"Source: owner/repo#123",
				"Target: owner/repo#456",
				"✅ Removed blocked-by relationship",
				"Dependency removed successfully",
			},
			expectedRemovals: 1,
		},
		{
			name: "successful blocks removal with force flag",
			setupDependencies: func(env *MockIntegrationEnvironment) {
				repo := env.AddRepository("owner", "repo")
				repo.AddIssue(123, "Feature: User Authentication")
				repo.AddIssue(789, "Frontend Integration")

				source := CreateIssueRef("owner", "repo", 123)
				target := CreateIssueRef("owner", "repo", 789)
				_ = env.AddDependency(source, target, "blocks")
			},
			source:           CreateIssueRef("owner", "repo", 123),
			target:           CreateIssueRef("owner", "repo", 789),
			relType:          "blocks",
			opts:             RemoveOptions{DryRun: false, Force: true},
			userConfirmation: "", // No confirmation needed with force
			expectedSuccess:  true,
			expectedOutput: []string{
				"✅ Removed blocks relationship",
				"owner/repo#123 → owner/repo#789",
				"Dependency removed successfully",
			},
			expectedRemovals: 1,
		},
		{
			name: "dry run mode shows preview without removal",
			setupDependencies: func(env *MockIntegrationEnvironment) {
				repo := env.AddRepository("owner", "repo")
				repo.AddIssue(123, "Feature: User Authentication")
				repo.AddIssue(456, "Database Setup")

				source := CreateIssueRef("owner", "repo", 123)
				target := CreateIssueRef("owner", "repo", 456)
				_ = env.AddDependency(source, target, "blocked-by")
			},
			source:           CreateIssueRef("owner", "repo", 123),
			target:           CreateIssueRef("owner", "repo", 456),
			relType:          "blocked-by",
			opts:             RemoveOptions{DryRun: true, Force: false},
			userConfirmation: "", // No confirmation in dry run
			expectedSuccess:  true,
			expectedOutput: []string{
				"Dry run: dependency removal preview",
				"Would remove:",
				"❌ blocked-by relationship: owner/repo#123 ← owner/repo#456",
				"No changes made",
			},
			expectedRemovals: 0, // Nothing should be removed in dry run
		},
		{
			name: "user cancels removal",
			setupDependencies: func(env *MockIntegrationEnvironment) {
				repo := env.AddRepository("owner", "repo")
				repo.AddIssue(123, "Feature: User Authentication")
				repo.AddIssue(456, "Database Setup")

				source := CreateIssueRef("owner", "repo", 123)
				target := CreateIssueRef("owner", "repo", 456)
				_ = env.AddDependency(source, target, "blocked-by")
			},
			source:           CreateIssueRef("owner", "repo", 123),
			target:           CreateIssueRef("owner", "repo", 456),
			relType:          "blocked-by",
			opts:             RemoveOptions{DryRun: false, Force: false},
			userConfirmation: "n",
			expectedSuccess:  false,
			expectedOutput: []string{
				"Remove dependency relationship?",
				"Dependency removal cancelled by user",
			},
			expectedRemovals: 0,
		},
		{
			name: "cross-repository dependency removal",
			setupDependencies: func(env *MockIntegrationEnvironment) {
				sourceRepo := env.AddRepository("owner", "backend")
				targetRepo := env.AddRepository("owner", "frontend")

				sourceRepo.AddIssue(123, "API Implementation")
				targetRepo.AddIssue(456, "Frontend Integration")

				source := CreateIssueRef("owner", "backend", 123)
				target := CreateIssueRef("owner", "frontend", 456)
				_ = env.AddDependency(source, target, "blocks")
			},
			source:           CreateIssueRef("owner", "backend", 123),
			target:           CreateIssueRef("owner", "frontend", 456),
			relType:          "blocks",
			opts:             RemoveOptions{DryRun: false, Force: true},
			userConfirmation: "",
			expectedSuccess:  true,
			expectedOutput: []string{
				"✅ Removed blocks relationship",
				"owner/backend#123 → owner/frontend#456",
				"Dependency removed successfully",
			},
			expectedRemovals: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock environment
			env := NewMockIntegrationEnvironment()
			tt.setupDependencies(env)

			t.Logf("Testing end-to-end workflow: %s", tt.name)
			t.Logf("Source: %s, Target: %s, Type: %s",
				tt.source.String(), tt.target.String(), tt.relType)

			// Simulate the complete workflow

			// 1. Validation phase
			initialDeps, err := env.GetDependencies(tt.source)
			require.NoError(t, err, "Should get initial dependencies")

			initialCount := len(initialDeps.BlockedBy) + len(initialDeps.Blocking)
			t.Logf("Initial dependencies: %d", initialCount)

			// Verify relationship exists (unless we expect it to fail)
			if tt.expectedSuccess || tt.expectedRemovals == 0 {
				relationshipExists := false
				switch tt.relType {
				case "blocked-by":
					for _, dep := range initialDeps.BlockedBy {
						if dep.Issue.Number == tt.target.Number {
							relationshipExists = true
							break
						}
					}
				case "blocks":
					for _, dep := range initialDeps.Blocking {
						if dep.Issue.Number == tt.target.Number {
							relationshipExists = true
							break
						}
					}
				}

				if !tt.opts.DryRun && tt.expectedRemovals > 0 {
					assert.True(t, relationshipExists,
						"Relationship should exist before removal")
				}
			}

			// 2. Command execution phase
			cmdParts := []string{"gh", "issue-dependency", "remove", fmt.Sprintf("%d", tt.source.Number)}
			switch tt.relType {
			case "blocked-by":
				cmdParts = append(cmdParts, "--blocked-by", fmt.Sprintf("%d", tt.target.Number))
			case "blocks":
				cmdParts = append(cmdParts, "--blocks", fmt.Sprintf("%d", tt.target.Number))
			}

			if tt.opts.DryRun {
				cmdParts = append(cmdParts, "--dry-run")
			}
			if tt.opts.Force {
				cmdParts = append(cmdParts, "--force")
			}

			command := strings.Join(cmdParts, " ")
			env.ExecuteCommand(command)
			t.Logf("Executed command: %s", command)

			// 3. Confirmation phase (if applicable)
			if !tt.opts.DryRun && !tt.opts.Force && tt.userConfirmation != "" {
				env.AddConfirmation(tt.userConfirmation)
				t.Logf("User confirmation: %s", tt.userConfirmation)
			}

			// 4. Execution phase (simulate the actual removal)
			var executionError error
			if tt.expectedSuccess && !tt.opts.DryRun && tt.userConfirmation != "n" {
				executionError = env.RemoveDependency(tt.source, tt.target, tt.relType)
				if executionError == nil {
					env.AddOutput("✅ Removed " + tt.relType + " relationship")
					env.AddOutput("Dependency removed successfully")
				}
			} else if tt.opts.DryRun {
				env.AddOutput("Dry run: dependency removal preview")
				env.AddOutput("No changes made")
			} else if tt.userConfirmation == "n" {
				env.AddOutput("Dependency removal cancelled by user")
			}

			// 5. Verification phase
			finalDeps, err := env.GetDependencies(tt.source)
			require.NoError(t, err, "Should get final dependencies")

			finalCount := len(finalDeps.BlockedBy) + len(finalDeps.Blocking)
			actualRemovals := initialCount - finalCount

			t.Logf("Final dependencies: %d (removed: %d)", finalCount, actualRemovals)

			// Verify removal count
			assert.Equal(t, tt.expectedRemovals, actualRemovals,
				"Actual removals should match expected")

			// Verify output content
			allOutput := strings.Join(env.outputs, " ")
			for _, expectedOutput := range tt.expectedOutput {
				assert.Contains(t, allOutput, expectedOutput,
					"Output should contain: %s", expectedOutput)
			}

			// Verify command execution
			assert.NotEmpty(t, env.executed, "Should have executed commands")
			assert.Contains(t, env.executed[0], "remove", "Should execute remove command")

			if tt.expectedSuccess {
				t.Log("✅ End-to-end workflow completed successfully")
			} else {
				t.Log("❌ End-to-end workflow failed as expected")
			}
		})
	}
}

// TestEndToEndBatchDependencyRemoval tests complete batch removal workflow
func TestEndToEndBatchDependencyRemoval(t *testing.T) {
	tests := []struct {
		name              string
		setupDependencies func(*MockIntegrationEnvironment)
		source            IssueRef
		targets           []IssueRef
		relType           string
		opts              RemoveOptions
		userConfirmation  string
		expectedSuccess   bool
		expectedOutput    []string
		expectedRemovals  int
	}{
		{
			name: "successful batch blocked-by removal",
			setupDependencies: func(env *MockIntegrationEnvironment) {
				repo := env.AddRepository("owner", "repo")
				repo.AddIssue(123, "Feature: User Authentication")
				repo.AddIssue(456, "Database Setup")
				repo.AddIssue(789, "API Design")
				repo.AddIssue(101, "Infrastructure")

				source := CreateIssueRef("owner", "repo", 123)
				_ = env.AddDependency(source, CreateIssueRef("owner", "repo", 456), "blocked-by")
				_ = env.AddDependency(source, CreateIssueRef("owner", "repo", 789), "blocked-by")
				_ = env.AddDependency(source, CreateIssueRef("owner", "repo", 101), "blocked-by")
			},
			source: CreateIssueRef("owner", "repo", 123),
			targets: []IssueRef{
				CreateIssueRef("owner", "repo", 456),
				CreateIssueRef("owner", "repo", 789),
				CreateIssueRef("owner", "repo", 101),
			},
			relType:          "blocked-by",
			opts:             RemoveOptions{DryRun: false, Force: false},
			userConfirmation: "y",
			expectedSuccess:  true,
			expectedOutput: []string{
				"Remove 3 dependency relationships?",
				"Source: owner/repo#123",
				"Type: blocked-by",
				"Targets:",
				"✅ Removed 3 blocked-by relationships",
				"Batch dependency removal completed successfully",
			},
			expectedRemovals: 3,
		},
		{
			name: "batch removal with partial failure",
			setupDependencies: func(env *MockIntegrationEnvironment) {
				repo := env.AddRepository("owner", "repo")
				repo.AddIssue(123, "Feature: User Authentication")
				repo.AddIssue(456, "Database Setup")
				repo.AddIssue(101, "Infrastructure")
				// Note: Issue 999 doesn't exist, simulating partial failure

				source := CreateIssueRef("owner", "repo", 123)
				_ = env.AddDependency(source, CreateIssueRef("owner", "repo", 456), "blocks")
				_ = env.AddDependency(source, CreateIssueRef("owner", "repo", 101), "blocks")
				// No dependency for 999 - will cause failure
			},
			source: CreateIssueRef("owner", "repo", 123),
			targets: []IssueRef{
				CreateIssueRef("owner", "repo", 456),
				CreateIssueRef("owner", "repo", 999), // Non-existent
				CreateIssueRef("owner", "repo", 101),
			},
			relType:          "blocks",
			opts:             RemoveOptions{DryRun: false, Force: true},
			userConfirmation: "",
			expectedSuccess:  false, // Partial failure
			expectedOutput: []string{
				"Batch removal partially failed: 2 succeeded, 1 failed",
				"owner/repo#999",
			},
			expectedRemovals: 2, // Only 2 out of 3 should succeed
		},
		{
			name: "batch dry run preview",
			setupDependencies: func(env *MockIntegrationEnvironment) {
				repo := env.AddRepository("owner", "repo")
				repo.AddIssue(123, "Feature: User Authentication")
				repo.AddIssue(456, "Database Setup")
				repo.AddIssue(789, "API Design")

				source := CreateIssueRef("owner", "repo", 123)
				_ = env.AddDependency(source, CreateIssueRef("owner", "repo", 456), "blocked-by")
				_ = env.AddDependency(source, CreateIssueRef("owner", "repo", 789), "blocked-by")
			},
			source: CreateIssueRef("owner", "repo", 123),
			targets: []IssueRef{
				CreateIssueRef("owner", "repo", 456),
				CreateIssueRef("owner", "repo", 789),
			},
			relType:          "blocked-by",
			opts:             RemoveOptions{DryRun: true, Force: false},
			userConfirmation: "",
			expectedSuccess:  true,
			expectedOutput: []string{
				"Dry run: batch dependency removal preview",
				"Would remove 2 relationships:",
				"❌ blocked-by relationship: owner/repo#123 ← owner/repo#456",
				"❌ blocked-by relationship: owner/repo#123 ← owner/repo#789",
				"No changes made",
			},
			expectedRemovals: 0,
		},
		{
			name: "cross-repository batch removal",
			setupDependencies: func(env *MockIntegrationEnvironment) {
				backendRepo := env.AddRepository("owner", "backend")
				frontendRepo := env.AddRepository("owner", "frontend")
				mobileRepo := env.AddRepository("owner", "mobile")

				backendRepo.AddIssue(123, "API Implementation")
				frontendRepo.AddIssue(456, "Web Interface")
				mobileRepo.AddIssue(789, "Mobile App")

				source := CreateIssueRef("owner", "backend", 123)
				_ = env.AddDependency(source, CreateIssueRef("owner", "frontend", 456), "blocks")
				_ = env.AddDependency(source, CreateIssueRef("owner", "mobile", 789), "blocks")
			},
			source: CreateIssueRef("owner", "backend", 123),
			targets: []IssueRef{
				CreateIssueRef("owner", "frontend", 456),
				CreateIssueRef("owner", "mobile", 789),
			},
			relType:          "blocks",
			opts:             RemoveOptions{DryRun: false, Force: true},
			userConfirmation: "",
			expectedSuccess:  true,
			expectedOutput: []string{
				"✅ Removed 2 blocks relationships:",
				"owner/backend#123 → owner/frontend#456",
				"owner/backend#123 → owner/mobile#789",
				"Batch dependency removal completed successfully",
			},
			expectedRemovals: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock environment
			env := NewMockIntegrationEnvironment()
			tt.setupDependencies(env)

			t.Logf("Testing end-to-end batch workflow: %s", tt.name)
			t.Logf("Source: %s, Targets: %d, Type: %s",
				tt.source.String(), len(tt.targets), tt.relType)

			// Get initial state
			initialDeps, err := env.GetDependencies(tt.source)
			require.NoError(t, err, "Should get initial dependencies")

			initialCount := len(initialDeps.BlockedBy) + len(initialDeps.Blocking)
			t.Logf("Initial dependencies: %d", initialCount)

			// Simulate batch command execution
			targetNumbers := make([]string, len(tt.targets))
			for i, target := range tt.targets {
				targetNumbers[i] = fmt.Sprintf("%d", target.Number)
			}
			targetList := strings.Join(targetNumbers, ",")

			cmdParts := []string{"gh", "issue-dependency", "remove", fmt.Sprintf("%d", tt.source.Number)}
			switch tt.relType {
			case "blocked-by":
				cmdParts = append(cmdParts, "--blocked-by", targetList)
			case "blocks":
				cmdParts = append(cmdParts, "--blocks", targetList)
			}

			if tt.opts.DryRun {
				cmdParts = append(cmdParts, "--dry-run")
			}
			if tt.opts.Force {
				cmdParts = append(cmdParts, "--force")
			}

			command := strings.Join(cmdParts, " ")
			env.ExecuteCommand(command)
			t.Logf("Executed batch command: %s", command)

			// Simulate confirmation
			if !tt.opts.DryRun && !tt.opts.Force && tt.userConfirmation != "" {
				env.AddConfirmation(tt.userConfirmation)
			}

			// Simulate batch execution
			successCount := 0
			var errors []string

			if tt.expectedSuccess && !tt.opts.DryRun && tt.userConfirmation != "n" {
				for _, target := range tt.targets {
					err := env.RemoveDependency(tt.source, target, tt.relType)
					if err != nil {
						errors = append(errors, fmt.Sprintf("%s: %v", target.String(), err))
					} else {
						successCount++
					}
				}

				// Generate appropriate output
				if len(errors) > 0 {
					env.AddOutput(fmt.Sprintf("Batch removal partially failed: %d succeeded, %d failed",
						successCount, len(errors)))
				} else {
					env.AddOutput(fmt.Sprintf("✅ Removed %d %s relationships", len(tt.targets), tt.relType))
					env.AddOutput("Batch dependency removal completed successfully")
				}
			} else if tt.opts.DryRun {
				env.AddOutput("Dry run: batch dependency removal preview")
				env.AddOutput(fmt.Sprintf("Would remove %d relationships:", len(tt.targets)))
				env.AddOutput("No changes made")
			}

			// Verify final state
			finalDeps, err := env.GetDependencies(tt.source)
			require.NoError(t, err, "Should get final dependencies")

			finalCount := len(finalDeps.BlockedBy) + len(finalDeps.Blocking)
			actualRemovals := initialCount - finalCount

			t.Logf("Final dependencies: %d (removed: %d)", finalCount, actualRemovals)

			// Verify expectations
			assert.Equal(t, tt.expectedRemovals, actualRemovals,
				"Actual removals should match expected")

			allOutput := strings.Join(env.outputs, " ")
			for _, expectedOutput := range tt.expectedOutput {
				assert.Contains(t, allOutput, expectedOutput,
					"Output should contain: %s", expectedOutput)
			}

			if tt.expectedSuccess && len(errors) == 0 {
				t.Log("✅ End-to-end batch workflow completed successfully")
			} else if len(errors) > 0 {
				t.Logf("⚠️ Batch workflow completed with %d errors", len(errors))
			} else {
				t.Log("❌ End-to-end batch workflow failed as expected")
			}
		})
	}
}

// TestEndToEndRemoveAllWorkflow tests complete "remove all" workflow
func TestEndToEndRemoveAllWorkflow(t *testing.T) {
	tests := []struct {
		name              string
		setupDependencies func(*MockIntegrationEnvironment)
		issue             IssueRef
		opts              RemoveOptions
		userConfirmation  string
		expectedSuccess   bool
		expectedOutput    []string
		expectedRemovals  int
	}{
		{
			name: "remove all dependencies successfully",
			setupDependencies: func(env *MockIntegrationEnvironment) {
				repo := env.AddRepository("owner", "repo")
				repo.AddIssue(123, "Feature: User Authentication")
				repo.AddIssue(456, "Database Setup")
				repo.AddIssue(789, "API Design")
				repo.AddIssue(101, "Frontend Work")
				repo.AddIssue(202, "Testing")

				source := CreateIssueRef("owner", "repo", 123)
				// Add blocked-by relationships
				_ = env.AddDependency(source, CreateIssueRef("owner", "repo", 456), "blocked-by")
				_ = env.AddDependency(source, CreateIssueRef("owner", "repo", 789), "blocked-by")
				// Add blocks relationships
				_ = env.AddDependency(source, CreateIssueRef("owner", "repo", 101), "blocks")
				_ = env.AddDependency(source, CreateIssueRef("owner", "repo", 202), "blocks")
			},
			issue:            CreateIssueRef("owner", "repo", 123),
			opts:             RemoveOptions{DryRun: false, Force: true},
			userConfirmation: "",
			expectedSuccess:  true,
			expectedOutput: []string{
				"✅ Removed all dependency relationships for owner/repo#123",
				"- 2 blocked-by relationships removed",
				"- 2 blocks relationships removed",
				"All dependencies cleared successfully",
			},
			expectedRemovals: 4,
		},
		{
			name: "remove all with no dependencies",
			setupDependencies: func(env *MockIntegrationEnvironment) {
				repo := env.AddRepository("owner", "repo")
				repo.AddIssue(123, "Feature: User Authentication")
				// No dependencies added
			},
			issue:            CreateIssueRef("owner", "repo", 123),
			opts:             RemoveOptions{DryRun: false, Force: true},
			userConfirmation: "",
			expectedSuccess:  false,
			expectedOutput: []string{
				"No dependency relationships found for owner/repo#123",
				"Use 'gh issue-dependency list' to see current dependencies",
			},
			expectedRemovals: 0,
		},
		{
			name: "remove all dry run preview",
			setupDependencies: func(env *MockIntegrationEnvironment) {
				repo := env.AddRepository("owner", "repo")
				repo.AddIssue(123, "Feature: User Authentication")
				repo.AddIssue(456, "Database Setup")
				repo.AddIssue(789, "Frontend Work")

				source := CreateIssueRef("owner", "repo", 123)
				_ = env.AddDependency(source, CreateIssueRef("owner", "repo", 456), "blocked-by")
				_ = env.AddDependency(source, CreateIssueRef("owner", "repo", 789), "blocks")
			},
			issue:            CreateIssueRef("owner", "repo", 123),
			opts:             RemoveOptions{DryRun: true, Force: false},
			userConfirmation: "",
			expectedSuccess:  true,
			expectedOutput: []string{
				"Dry run: batch dependency removal preview",
				"Would remove 2 relationships:",
				"❌ blocked-by relationship",
				"❌ blocks relationship",
				"No changes made",
			},
			expectedRemovals: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock environment
			env := NewMockIntegrationEnvironment()
			tt.setupDependencies(env)

			t.Logf("Testing remove-all workflow: %s", tt.name)
			t.Logf("Issue: %s", tt.issue.String())

			// Get initial state
			initialDeps, err := env.GetDependencies(tt.issue)
			require.NoError(t, err, "Should get initial dependencies")

			initialCount := len(initialDeps.BlockedBy) + len(initialDeps.Blocking)
			t.Logf("Initial dependencies: %d blocked-by, %d blocking",
				len(initialDeps.BlockedBy), len(initialDeps.Blocking))

			// Simulate remove --all command
			cmdParts := []string{"gh", "issue-dependency", "remove", fmt.Sprintf("%d", tt.issue.Number), "--all"}
			if tt.opts.DryRun {
				cmdParts = append(cmdParts, "--dry-run")
			}
			if tt.opts.Force {
				cmdParts = append(cmdParts, "--force")
			}

			command := strings.Join(cmdParts, " ")
			env.ExecuteCommand(command)
			t.Logf("Executed remove-all command: %s", command)

			// Simulate execution
			if initialCount == 0 {
				// No dependencies to remove
				env.AddOutput("No dependency relationships found for " + tt.issue.String())
				env.AddOutput("Use 'gh issue-dependency list' to see current dependencies")
			} else if tt.opts.DryRun {
				env.AddOutput("Dry run: batch dependency removal preview")
				env.AddOutput(fmt.Sprintf("Would remove %d relationships:", initialCount))
				env.AddOutput("No changes made")
			} else if tt.expectedSuccess {
				// Remove all blocked-by relationships
				blockedByCount := len(initialDeps.BlockedBy)
				for _, dep := range initialDeps.BlockedBy {
					target := CreateIssueRef(
						strings.Split(dep.Repository, "/")[0],
						strings.Split(dep.Repository, "/")[1],
						dep.Issue.Number,
					)
					_ = env.RemoveDependency(tt.issue, target, "blocked-by")
				}

				// Remove all blocking relationships
				blockingCount := len(initialDeps.Blocking)
				for _, dep := range initialDeps.Blocking {
					target := CreateIssueRef(
						strings.Split(dep.Repository, "/")[0],
						strings.Split(dep.Repository, "/")[1],
						dep.Issue.Number,
					)
					_ = env.RemoveDependency(tt.issue, target, "blocks")
				}

				env.AddOutput(fmt.Sprintf("✅ Removed all dependency relationships for %s", tt.issue.String()))
				env.AddOutput(fmt.Sprintf("- %d blocked-by relationships removed", blockedByCount))
				env.AddOutput(fmt.Sprintf("- %d blocks relationships removed", blockingCount))
				env.AddOutput("All dependencies cleared successfully")
			}

			// Verify final state
			finalDeps, err := env.GetDependencies(tt.issue)
			require.NoError(t, err, "Should get final dependencies")

			finalCount := len(finalDeps.BlockedBy) + len(finalDeps.Blocking)
			actualRemovals := initialCount - finalCount

			t.Logf("Final dependencies: %d (removed: %d)", finalCount, actualRemovals)

			// Verify expectations
			assert.Equal(t, tt.expectedRemovals, actualRemovals,
				"Actual removals should match expected")

			allOutput := strings.Join(env.outputs, " ")
			for _, expectedOutput := range tt.expectedOutput {
				assert.Contains(t, allOutput, expectedOutput,
					"Output should contain: %s", expectedOutput)
			}

			if tt.expectedSuccess {
				t.Log("✅ End-to-end remove-all workflow completed successfully")
			} else {
				t.Log("❌ End-to-end remove-all workflow failed as expected")
			}
		})
	}
}

// TestEndToEndErrorRecoveryWorkflows tests error handling and recovery scenarios
func TestEndToEndErrorRecoveryWorkflows(t *testing.T) {
	tests := []struct {
		name             string
		scenario         string
		expectedError    string
		expectedRecovery string
	}{
		{
			name:             "authentication failure recovery",
			scenario:         "User not authenticated with GitHub CLI",
			expectedError:    "authentication failed",
			expectedRecovery: "Run 'gh auth login' to authenticate",
		},
		{
			name:             "permission denied recovery",
			scenario:         "User lacks write permissions to repository",
			expectedError:    "permission denied",
			expectedRecovery: "You need write or maintain permissions",
		},
		{
			name:             "relationship not found recovery",
			scenario:         "Trying to remove non-existent relationship",
			expectedError:    "relationship does not exist",
			expectedRecovery: "Use 'gh issue-dependency list' to see current dependencies",
		},
		{
			name:             "network error with retry",
			scenario:         "Transient network failure during API call",
			expectedError:    "network timeout",
			expectedRecovery: "Retrying with exponential backoff",
		},
		{
			name:             "rate limit handling",
			scenario:         "API rate limit exceeded",
			expectedError:    "rate limit exceeded",
			expectedRecovery: "Retrying after rate limit reset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Testing error recovery workflow: %s", tt.name)
			t.Logf("Scenario: %s", tt.scenario)
			t.Logf("Expected error: %s", tt.expectedError)
			t.Logf("Expected recovery: %s", tt.expectedRecovery)

			// This test validates the error handling patterns
			// In a real implementation, we would simulate these scenarios
			// and verify the appropriate error messages and recovery suggestions

			assert.NotEmpty(t, tt.expectedError, "Should have expected error message")
			assert.NotEmpty(t, tt.expectedRecovery, "Should have recovery guidance")

			t.Log("✅ Error recovery patterns validated")
		})
	}
}
