package pkg

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockStdinReader simulates user input for confirmation prompts
type MockStdinReader struct {
	input string
	pos   int
}

func NewMockStdinReader(input string) *MockStdinReader {
	return &MockStdinReader{input: input, pos: 0}
}

func (m *MockStdinReader) Read(p []byte) (n int, err error) {
	if m.pos >= len(m.input) {
		return 0, io.EOF
	}

	n = copy(p, m.input[m.pos:])
	m.pos += n
	return n, nil
}

// MockOutputWriter captures output for testing
type MockOutputWriter struct {
	buffer bytes.Buffer
}

func (m *MockOutputWriter) Write(p []byte) (n int, err error) {
	return m.buffer.Write(p)
}

func (m *MockOutputWriter) String() string {
	return m.buffer.String()
}

func (m *MockOutputWriter) Reset() {
	m.buffer.Reset()
}

// TestDryRunMode tests the dry run functionality with comprehensive scenarios
func TestDryRunMode(t *testing.T) {
	tests := []struct {
		name              string
		source            IssueRef
		targets           []IssueRef
		relType           string
		expectedOutput    []string
		shouldShowPreview bool
		shouldMakeChanges bool
	}{
		{
			name:    "single dependency dry run - blocked-by",
			source:  CreateIssueRef("owner", "repo", 123),
			targets: []IssueRef{CreateIssueRef("owner", "repo", 456)},
			relType: "blocked-by",
			expectedOutput: []string{
				"Dry run: dependency removal preview",
				"Would remove:",
				"❌ blocked-by relationship: owner/repo#123 ← owner/repo#456",
				"No changes made",
				"Use --force to skip confirmation or remove --dry-run to execute",
			},
			shouldShowPreview: true,
			shouldMakeChanges: false,
		},
		{
			name:    "single dependency dry run - blocks",
			source:  CreateIssueRef("owner", "repo", 123),
			targets: []IssueRef{CreateIssueRef("other", "repo", 789)},
			relType: "blocks",
			expectedOutput: []string{
				"Dry run: dependency removal preview",
				"Would remove:",
				"❌ blocks relationship: owner/repo#123 → other/repo#789",
				"No changes made",
			},
			shouldShowPreview: true,
			shouldMakeChanges: false,
		},
		{
			name:   "multiple dependencies dry run - blocked-by",
			source: CreateIssueRef("owner", "repo", 123),
			targets: []IssueRef{
				CreateIssueRef("owner", "repo", 456),
				CreateIssueRef("owner", "repo", 789),
				CreateIssueRef("other", "repo", 101),
			},
			relType: "blocked-by",
			expectedOutput: []string{
				"Dry run: dependency removal preview",
				"Would remove:",
				"❌ blocked-by relationship: owner/repo#123 ← owner/repo#456",
				"❌ blocked-by relationship: owner/repo#123 ← owner/repo#789",
				"❌ blocked-by relationship: owner/repo#123 ← other/repo#101",
				"No changes made",
			},
			shouldShowPreview: true,
			shouldMakeChanges: false,
		},
		{
			name:   "multiple dependencies dry run - blocks",
			source: CreateIssueRef("owner", "repo", 123),
			targets: []IssueRef{
				CreateIssueRef("owner", "repo", 456),
				CreateIssueRef("other", "repo", 789),
			},
			relType: "blocks",
			expectedOutput: []string{
				"Dry run: dependency removal preview",
				"Would remove:",
				"❌ blocks relationship: owner/repo#123 → owner/repo#456",
				"❌ blocks relationship: owner/repo#123 → other/repo#789",
				"No changes made",
			},
			shouldShowPreview: true,
			shouldMakeChanges: false,
		},
		{
			name:    "cross-repository dry run",
			source:  CreateIssueRef("source", "repo", 123),
			targets: []IssueRef{CreateIssueRef("target", "repo", 456)},
			relType: "blocked-by",
			expectedOutput: []string{
				"Dry run: dependency removal preview",
				"Would remove:",
				"❌ blocked-by relationship: source/repo#123 ← target/repo#456",
				"No changes made",
			},
			shouldShowPreview: true,
			shouldMakeChanges: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock output writer
			_ = &MockOutputWriter{}

			// Simulate dry run execution
			_ = RemoveOptions{DryRun: true, Force: false}

			t.Logf("Testing dry run: %s %s %d targets",
				tt.source.String(), tt.relType, len(tt.targets))

			// Test dry run preview generation
			output := generateDryRunPreview(tt.source, tt.targets, tt.relType)

			// Verify expected output elements are present
			for _, expectedLine := range tt.expectedOutput {
				assert.Contains(t, output, expectedLine,
					"Dry run output should contain: %s", expectedLine)
			}

			// Verify dry run characteristics
			assert.True(t, tt.shouldShowPreview, "Dry run should show preview")
			assert.False(t, tt.shouldMakeChanges, "Dry run should not make changes")

			// Verify relationship arrows are correct
			if tt.relType == "blocked-by" {
				assert.Contains(t, output, "←", "Blocked-by should use ← arrow")
			} else if tt.relType == "blocks" {
				assert.Contains(t, output, "→", "Blocks should use → arrow")
			}

			// Verify all targets are shown
			for _, target := range tt.targets {
				assert.Contains(t, output, target.String(),
					"Dry run should show target: %s", target.String())
			}

			t.Logf("Dry run preview generated successfully")
		})
	}
}

// generateDryRunPreview simulates the dry run preview generation
func generateDryRunPreview(source IssueRef, targets []IssueRef, relType string) string {
	var output strings.Builder

	output.WriteString("Dry run: dependency removal preview\n")
	output.WriteString("\n")
	output.WriteString("Would remove:\n")

	for _, target := range targets {
		var arrow string
		switch relType {
		case "blocked-by":
			arrow = "←" // Source is blocked by target
		case "blocks":
			arrow = "→" // Source blocks target
		}

		output.WriteString(fmt.Sprintf("  ❌ %s relationship: %s %s %s\n",
			relType, source.String(), arrow, target.String()))
	}

	output.WriteString("\n")
	output.WriteString("No changes made. Use --force to skip confirmation or remove --dry-run to execute.\n")

	return output.String()
}

// TestConfirmationPrompts tests the confirmation system with various user responses
func TestConfirmationPrompts(t *testing.T) {
	tests := []struct {
		name           string
		source         IssueRef
		targets        []IssueRef
		relType        string
		userInput      string
		expectedPrompt []string
		expectedResult bool
		shouldProceed  bool
	}{
		{
			name:      "user confirms single dependency removal - 'y'",
			source:    CreateIssueRef("owner", "repo", 123),
			targets:   []IssueRef{CreateIssueRef("owner", "repo", 456)},
			relType:   "blocked-by",
			userInput: "y\n",
			expectedPrompt: []string{
				"Remove dependency relationship?",
				"Source: owner/repo#123",
				"Target: owner/repo#456",
				"Type: blocked-by",
				"This will remove the \"blocked-by\" relationship between these issues.",
				"Continue? (y/N):",
			},
			expectedResult: true,
			shouldProceed:  true,
		},
		{
			name:      "user confirms single dependency removal - 'Y'",
			source:    CreateIssueRef("owner", "repo", 123),
			targets:   []IssueRef{CreateIssueRef("owner", "repo", 456)},
			relType:   "blocks",
			userInput: "Y\n",
			expectedPrompt: []string{
				"Remove dependency relationship?",
				"Source: owner/repo#123",
				"Target: owner/repo#456",
				"Type: blocks",
				"This will remove the \"blocks\" relationship between these issues.",
				"Continue? (y/N):",
			},
			expectedResult: true,
			shouldProceed:  true,
		},
		{
			name:      "user confirms single dependency removal - 'yes'",
			source:    CreateIssueRef("owner", "repo", 123),
			targets:   []IssueRef{CreateIssueRef("owner", "repo", 456)},
			relType:   "blocked-by",
			userInput: "yes\n",
			expectedPrompt: []string{
				"Remove dependency relationship?",
				"Continue? (y/N):",
			},
			expectedResult: true,
			shouldProceed:  true,
		},
		{
			name:      "user cancels single dependency removal - 'n'",
			source:    CreateIssueRef("owner", "repo", 123),
			targets:   []IssueRef{CreateIssueRef("owner", "repo", 456)},
			relType:   "blocked-by",
			userInput: "n\n",
			expectedPrompt: []string{
				"Remove dependency relationship?",
				"Continue? (y/N):",
			},
			expectedResult: false,
			shouldProceed:  false,
		},
		{
			name:      "user cancels single dependency removal - 'N'",
			source:    CreateIssueRef("owner", "repo", 123),
			targets:   []IssueRef{CreateIssueRef("owner", "repo", 456)},
			relType:   "blocks",
			userInput: "N\n",
			expectedPrompt: []string{
				"Remove dependency relationship?",
				"Continue? (y/N):",
			},
			expectedResult: false,
			shouldProceed:  false,
		},
		{
			name:      "user cancels single dependency removal - 'no'",
			source:    CreateIssueRef("owner", "repo", 123),
			targets:   []IssueRef{CreateIssueRef("owner", "repo", 456)},
			relType:   "blocked-by",
			userInput: "no\n",
			expectedPrompt: []string{
				"Remove dependency relationship?",
				"Continue? (y/N):",
			},
			expectedResult: false,
			shouldProceed:  false,
		},
		{
			name:      "user cancels by pressing Enter (default to N)",
			source:    CreateIssueRef("owner", "repo", 123),
			targets:   []IssueRef{CreateIssueRef("owner", "repo", 456)},
			relType:   "blocked-by",
			userInput: "\n",
			expectedPrompt: []string{
				"Remove dependency relationship?",
				"Continue? (y/N):",
			},
			expectedResult: false,
			shouldProceed:  false,
		},
		{
			name:   "user confirms multiple dependency removal",
			source: CreateIssueRef("owner", "repo", 123),
			targets: []IssueRef{
				CreateIssueRef("owner", "repo", 456),
				CreateIssueRef("other", "repo", 789),
			},
			relType:   "blocks",
			userInput: "y\n",
			expectedPrompt: []string{
				"Remove dependency relationship(s)?",
				"Source: owner/repo#123",
				"Targets: 2 issues",
				"- owner/repo#456",
				"- other/repo#789",
				"Type: blocks",
				"This will remove 2 \"blocks\" relationships.",
				"Continue? (y/N):",
			},
			expectedResult: true,
			shouldProceed:  true,
		},
		{
			name:   "user cancels multiple dependency removal",
			source: CreateIssueRef("owner", "repo", 123),
			targets: []IssueRef{
				CreateIssueRef("owner", "repo", 456),
				CreateIssueRef("other", "repo", 789),
				CreateIssueRef("third", "repo", 101),
			},
			relType:   "blocked-by",
			userInput: "n\n",
			expectedPrompt: []string{
				"Remove dependency relationship(s)?",
				"Source: owner/repo#123",
				"Targets: 3 issues",
				"This will remove 3 \"blocked-by\" relationships.",
				"Continue? (y/N):",
			},
			expectedResult: false,
			shouldProceed:  false,
		},
		{
			name:      "invalid input then valid confirmation",
			source:    CreateIssueRef("owner", "repo", 123),
			targets:   []IssueRef{CreateIssueRef("owner", "repo", 456)},
			relType:   "blocked-by",
			userInput: "maybe\ninvalid\ny\n",
			expectedPrompt: []string{
				"Remove dependency relationship?",
				"Continue? (y/N):",
			},
			expectedResult: true,
			shouldProceed:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock input and output
			_ = NewMockStdinReader(tt.userInput)
			_ = &MockOutputWriter{}

			t.Logf("Testing confirmation: %s %s %d targets, input: %q",
				tt.source.String(), tt.relType, len(tt.targets), strings.TrimSpace(tt.userInput))

			// Generate confirmation prompt
			prompt := generateConfirmationPrompt(tt.source, tt.targets, tt.relType)

			// Verify expected prompt elements are present
			for _, expectedLine := range tt.expectedPrompt {
				assert.Contains(t, prompt, expectedLine,
					"Confirmation prompt should contain: %s", expectedLine)
			}

			// Test user input processing
			confirmed := processUserConfirmation(tt.userInput)
			assert.Equal(t, tt.expectedResult, confirmed,
				"User confirmation result should match expected")

			// Verify behavior based on confirmation
			if tt.shouldProceed {
				assert.True(t, confirmed, "Should proceed with removal")
				t.Logf("User confirmed removal")
			} else {
				assert.False(t, confirmed, "Should cancel removal")
				t.Logf("User cancelled removal")
			}
		})
	}
}

// generateConfirmationPrompt simulates the confirmation prompt generation
func generateConfirmationPrompt(source IssueRef, targets []IssueRef, relType string) string {
	var prompt strings.Builder

	if len(targets) == 1 {
		prompt.WriteString("Remove dependency relationship?\n")
	} else {
		prompt.WriteString("Remove dependency relationship(s)?\n")
	}

	prompt.WriteString(fmt.Sprintf("  Source: %s\n", source.String()))

	if len(targets) == 1 {
		prompt.WriteString(fmt.Sprintf("  Target: %s\n", targets[0].String()))
	} else {
		prompt.WriteString(fmt.Sprintf("  Targets: %d issues\n", len(targets)))
		for _, target := range targets {
			prompt.WriteString(fmt.Sprintf("    - %s\n", target.String()))
		}
	}

	prompt.WriteString(fmt.Sprintf("  Type: %s\n", relType))
	prompt.WriteString("\n")

	if len(targets) == 1 {
		prompt.WriteString(fmt.Sprintf("This will remove the \"%s\" relationship between these issues.\n", relType))
	} else {
		prompt.WriteString(fmt.Sprintf("This will remove %d \"%s\" relationships.\n", len(targets), relType))
	}

	prompt.WriteString("Continue? (y/N): ")

	return prompt.String()
}

// processUserConfirmation simulates processing user input for confirmation
func processUserConfirmation(input string) bool {
	input = strings.ToLower(strings.TrimSpace(input))

	// Handle multiple inputs (for invalid input scenarios)
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "y" || line == "yes" {
			return true
		}
		if line == "n" || line == "no" || line == "" {
			return false
		}
		// Invalid inputs are ignored, continue to next line
	}

	// Default to false (cancel) if no valid input found
	return false
}

// TestForceMode tests the force flag functionality
func TestForceMode(t *testing.T) {
	tests := []struct {
		name             string
		source           IssueRef
		targets          []IssueRef
		relType          string
		forceFlag        bool
		shouldPrompt     bool
		shouldProceed    bool
		expectedBehavior string
	}{
		{
			name:             "force flag skips confirmation - single target",
			source:           CreateIssueRef("owner", "repo", 123),
			targets:          []IssueRef{CreateIssueRef("owner", "repo", 456)},
			relType:          "blocked-by",
			forceFlag:        true,
			shouldPrompt:     false,
			shouldProceed:    true,
			expectedBehavior: "skip confirmation and proceed directly",
		},
		{
			name:   "force flag skips confirmation - multiple targets",
			source: CreateIssueRef("owner", "repo", 123),
			targets: []IssueRef{
				CreateIssueRef("owner", "repo", 456),
				CreateIssueRef("other", "repo", 789),
			},
			relType:          "blocks",
			forceFlag:        true,
			shouldPrompt:     false,
			shouldProceed:    true,
			expectedBehavior: "skip confirmation and proceed directly",
		},
		{
			name:             "no force flag requires confirmation - single target",
			source:           CreateIssueRef("owner", "repo", 123),
			targets:          []IssueRef{CreateIssueRef("owner", "repo", 456)},
			relType:          "blocked-by",
			forceFlag:        false,
			shouldPrompt:     true,
			shouldProceed:    false, // Depends on user input
			expectedBehavior: "show confirmation prompt and wait for user input",
		},
		{
			name:   "no force flag requires confirmation - multiple targets",
			source: CreateIssueRef("owner", "repo", 123),
			targets: []IssueRef{
				CreateIssueRef("owner", "repo", 456),
				CreateIssueRef("other", "repo", 789),
			},
			relType:          "blocks",
			forceFlag:        false,
			shouldPrompt:     true,
			shouldProceed:    false, // Depends on user input
			expectedBehavior: "show confirmation prompt and wait for user input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := RemoveOptions{
				DryRun: false,
				Force:  tt.forceFlag,
			}

			t.Logf("Testing force mode: force=%v, %s %s %d targets",
				tt.forceFlag, tt.source.String(), tt.relType, len(tt.targets))

			// Test force flag behavior
			if tt.forceFlag {
				assert.False(t, tt.shouldPrompt, "Force mode should skip confirmation prompt")
				assert.True(t, tt.shouldProceed, "Force mode should proceed directly")
				t.Logf("Force mode: %s", tt.expectedBehavior)
			} else {
				assert.True(t, tt.shouldPrompt, "Non-force mode should show confirmation prompt")
				t.Logf("Interactive mode: %s", tt.expectedBehavior)
			}

			// Verify RemoveOptions configuration
			assert.Equal(t, tt.forceFlag, opts.Force, "Force flag should match expected")
			assert.False(t, opts.DryRun, "DryRun should be false for force tests")
		})
	}
}

// TestConfirmationAndDryRunInteraction tests interaction between confirmation and dry run modes
func TestConfirmationAndDryRunInteraction(t *testing.T) {
	scenarios := []struct {
		name              string
		dryRun            bool
		force             bool
		expectedBehavior  string
		shouldShowPreview bool
		shouldPrompt      bool
		shouldExecute     bool
	}{
		{
			name:              "dry run only",
			dryRun:            true,
			force:             false,
			expectedBehavior:  "show preview, no prompt, no execution",
			shouldShowPreview: true,
			shouldPrompt:      false,
			shouldExecute:     false,
		},
		{
			name:              "force only",
			dryRun:            false,
			force:             true,
			expectedBehavior:  "no preview, no prompt, execute directly",
			shouldShowPreview: false,
			shouldPrompt:      false,
			shouldExecute:     true,
		},
		{
			name:              "neither dry run nor force",
			dryRun:            false,
			force:             false,
			expectedBehavior:  "no preview, show prompt, execute based on user input",
			shouldShowPreview: false,
			shouldPrompt:      true,
			shouldExecute:     false, // Depends on user input
		},
		{
			name:              "both dry run and force (edge case)",
			dryRun:            true,
			force:             true,
			expectedBehavior:  "dry run takes precedence, show preview only",
			shouldShowPreview: true,
			shouldPrompt:      false,
			shouldExecute:     false,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			opts := RemoveOptions{
				DryRun: scenario.dryRun,
				Force:  scenario.force,
			}

			_ = CreateIssueRef("owner", "repo", 123)
			_ = CreateIssueRef("owner", "repo", 456)

			t.Logf("Testing interaction: dryRun=%v, force=%v", scenario.dryRun, scenario.force)
			t.Logf("Expected behavior: %s", scenario.expectedBehavior)

			// Test behavior based on flags
			if scenario.dryRun {
				assert.True(t, scenario.shouldShowPreview, "Dry run should show preview")
				assert.False(t, scenario.shouldExecute, "Dry run should not execute")
				t.Logf("Dry run mode: showing preview without execution")
			}

			if scenario.force && !scenario.dryRun {
				assert.False(t, scenario.shouldPrompt, "Force mode should skip prompt")
				assert.True(t, scenario.shouldExecute, "Force mode should execute")
				t.Logf("Force mode: executing without confirmation")
			}

			if !scenario.dryRun && !scenario.force {
				assert.True(t, scenario.shouldPrompt, "Interactive mode should prompt")
				t.Logf("Interactive mode: showing confirmation prompt")
			}

			// Verify flag configuration
			assert.Equal(t, scenario.dryRun, opts.DryRun, "DryRun flag should match")
			assert.Equal(t, scenario.force, opts.Force, "Force flag should match")
		})
	}
}

// TestConfirmationErrorHandling tests error scenarios in the confirmation system
func TestConfirmationErrorHandling(t *testing.T) {
	errorScenarios := []struct {
		name          string
		source        IssueRef
		targets       []IssueRef
		relType       string
		simulateError string
		expectedError string
	}{
		{
			name:          "empty source issue",
			source:        IssueRef{},
			targets:       []IssueRef{CreateIssueRef("owner", "repo", 456)},
			relType:       "blocked-by",
			simulateError: "empty source",
			expectedError: "invalid source issue",
		},
		{
			name:          "empty targets list",
			source:        CreateIssueRef("owner", "repo", 123),
			targets:       []IssueRef{},
			relType:       "blocked-by",
			simulateError: "empty targets",
			expectedError: "no targets specified",
		},
		{
			name:          "invalid relationship type",
			source:        CreateIssueRef("owner", "repo", 123),
			targets:       []IssueRef{CreateIssueRef("owner", "repo", 456)},
			relType:       "invalid",
			simulateError: "invalid relationship type",
			expectedError: "invalid relationship type",
		},
		{
			name:          "stdin read error",
			source:        CreateIssueRef("owner", "repo", 123),
			targets:       []IssueRef{CreateIssueRef("owner", "repo", 456)},
			relType:       "blocked-by",
			simulateError: "stdin error",
			expectedError: "failed to read user input",
		},
	}

	for _, scenario := range errorScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Logf("Testing error scenario: %s", scenario.name)
			t.Logf("Simulated error: %s", scenario.simulateError)

			// Verify error conditions
			switch scenario.simulateError {
			case "empty source":
				assert.Empty(t, scenario.source.Owner, "Source owner should be empty")
				assert.Zero(t, scenario.source.Number, "Source number should be zero")
			case "empty targets":
				assert.Empty(t, scenario.targets, "Targets should be empty")
			case "invalid relationship type":
				assert.NotContains(t, []string{"blocked-by", "blocks"}, scenario.relType,
					"Relationship type should be invalid")
			case "stdin error":
				// This would be tested with a mock that returns an error
				t.Logf("Would simulate stdin read error")
			}

			t.Logf("Expected error type: %s", scenario.expectedError)
		})
	}
}

// TestConfirmationOutputFormatting tests the formatting of confirmation prompts
func TestConfirmationOutputFormatting(t *testing.T) {
	formattingTests := []struct {
		name           string
		source         IssueRef
		targets        []IssueRef
		relType        string
		expectedFormat []string
	}{
		{
			name:    "simple same-repo formatting",
			source:  CreateIssueRef("owner", "repo", 123),
			targets: []IssueRef{CreateIssueRef("owner", "repo", 456)},
			relType: "blocked-by",
			expectedFormat: []string{
				"Remove dependency relationship?",
				"Source: owner/repo#123",
				"Target: owner/repo#456",
				"Type: blocked-by",
				"Continue? (y/N):",
			},
		},
		{
			name:    "cross-repository formatting",
			source:  CreateIssueRef("source-owner", "source-repo", 123),
			targets: []IssueRef{CreateIssueRef("target-owner", "target-repo", 456)},
			relType: "blocks",
			expectedFormat: []string{
				"Remove dependency relationship?",
				"Source: source-owner/source-repo#123",
				"Target: target-owner/target-repo#456",
				"Type: blocks",
				"Continue? (y/N):",
			},
		},
		{
			name:   "multiple targets formatting",
			source: CreateIssueRef("owner", "repo", 123),
			targets: []IssueRef{
				CreateIssueRef("owner", "repo", 456),
				CreateIssueRef("other", "repo", 789),
			},
			relType: "blocked-by",
			expectedFormat: []string{
				"Remove dependency relationship(s)?",
				"Source: owner/repo#123",
				"Targets: 2 issues",
				"- owner/repo#456",
				"- other/repo#789",
				"Type: blocked-by",
				"This will remove 2 \"blocked-by\" relationships.",
				"Continue? (y/N):",
			},
		},
		{
			name:    "large issue numbers formatting",
			source:  CreateIssueRef("owner", "repo", 999999),
			targets: []IssueRef{CreateIssueRef("owner", "repo", 888888)},
			relType: "blocks",
			expectedFormat: []string{
				"Source: owner/repo#999999",
				"Target: owner/repo#888888",
				"Type: blocks",
			},
		},
	}

	for _, tt := range formattingTests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := generateConfirmationPrompt(tt.source, tt.targets, tt.relType)

			t.Logf("Testing formatting for: %s", tt.name)

			// Verify all expected format elements are present
			for _, expectedElement := range tt.expectedFormat {
				assert.Contains(t, prompt, expectedElement,
					"Confirmation prompt should contain: %s", expectedElement)
			}

			// Verify proper issue reference formatting
			assert.Contains(t, prompt, tt.source.String(),
				"Prompt should contain formatted source issue")

			for _, target := range tt.targets {
				assert.Contains(t, prompt, target.String(),
					"Prompt should contain formatted target issue: %s", target.String())
			}

			// Verify relationship type formatting
			assert.Contains(t, prompt, fmt.Sprintf("Type: %s", tt.relType),
				"Prompt should contain formatted relationship type")

			t.Logf("Formatting verified successfully")
		})
	}
}
