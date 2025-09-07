package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCode int
		wantOut  string
		wantErr  string
	}{
		{
			name:     "no args shows help",
			args:     []string{},
			wantCode: 0,
			wantOut:  "A GitHub CLI extension for managing issue dependencies",
		},
		{
			name:     "help flag",
			args:     []string{"--help"},
			wantCode: 0,
			wantOut:  "A GitHub CLI extension for managing issue dependencies",
		},
		{
			name:     "version flag",
			args:     []string{"--version"},
			wantCode: 0,
			wantOut:  "gh-issue-dependency version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new root command for each test to avoid flag pollution
			cmd := createTestRootCmd()
			
			// Capture output
			outBuf := &bytes.Buffer{}
			errBuf := &bytes.Buffer{}
			cmd.SetOut(outBuf)
			cmd.SetErr(errBuf)
			
			// Set args
			cmd.SetArgs(tt.args)
			
			// Execute command
			err := cmd.Execute()
			
			// Check exit code
			if tt.wantCode == 0 {
				assert.NoError(t, err, "command should succeed")
			} else {
				assert.Error(t, err, "command should fail")
			}
			
			// Check output
			if tt.wantOut != "" {
				output := outBuf.String()
				if output == "" {
					// Sometimes help goes to stderr
					output = errBuf.String()
				}
				assert.Contains(t, output, tt.wantOut, "output should contain expected text")
			}
			
			// Check error output
			if tt.wantErr != "" {
				assert.Contains(t, errBuf.String(), tt.wantErr, "error output should contain expected text")
			}
		})
	}
}

func TestRootCmdConfiguration(t *testing.T) {
	cmd := createTestRootCmd()
	
	// Test command configuration
	assert.Equal(t, "gh-issue-dependency", cmd.Use)
	assert.Equal(t, "Manage issue dependencies in GitHub repositories", cmd.Short)
	assert.Contains(t, cmd.Long, "A GitHub CLI extension for managing issue dependencies")
	assert.Equal(t, Version, cmd.Version)
	
	// Test that examples are included
	assert.Contains(t, cmd.Long, "Examples:")
	assert.Contains(t, cmd.Long, "gh issue-dependency list")
	assert.Contains(t, cmd.Long, "gh issue-dependency add")
	assert.Contains(t, cmd.Long, "gh issue-dependency remove")
}

func TestRootCmdFlags(t *testing.T) {
	cmd := createTestRootCmd()
	
	// Test global --repo flag
	repoFlag := cmd.PersistentFlags().Lookup("repo")
	require.NotNil(t, repoFlag, "repo flag should exist")
	assert.Equal(t, "R", repoFlag.Shorthand, "repo flag should have -R shorthand")
	assert.Equal(t, "", repoFlag.DefValue, "repo flag should have empty default")
	assert.Contains(t, repoFlag.Usage, "[HOST/]OWNER/REPO", "repo flag should have correct usage text")
}

func TestVersionTemplate(t *testing.T) {
	cmd := createTestRootCmd()
	
	// Capture version output
	outBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)
	cmd.SetArgs([]string{"--version"})
	
	err := cmd.Execute()
	assert.NoError(t, err)
	
	output := outBuf.String()
	assert.Contains(t, output, "gh-issue-dependency version")
	assert.Contains(t, output, Version)
}

func TestExecuteFunction(t *testing.T) {
	tests := []struct {
		name     string
		mockCmd  func() *cobra.Command
		wantCode int
	}{
		{
			name: "successful execution",
			mockCmd: func() *cobra.Command {
				cmd := &cobra.Command{
					Use: "test",
					RunE: func(cmd *cobra.Command, args []string) error {
						return nil
					},
				}
				return cmd
			},
			wantCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original rootCmd
			originalRootCmd := rootCmd
			defer func() { rootCmd = originalRootCmd }()
			
			// Replace rootCmd with mock
			rootCmd = tt.mockCmd()
			
			// Test Execute function
			code := Execute()
			assert.Equal(t, tt.wantCode, code)
		})
	}
}

func TestRootCmdSubcommands(t *testing.T) {
	cmd := createTestRootCmd()
	
	// Test that subcommands can be added (this tests the structure)
	addCmd := &cobra.Command{Use: "add"}
	listCmd := &cobra.Command{Use: "list"}  
	removeCmd := &cobra.Command{Use: "remove"}
	
	cmd.AddCommand(addCmd, listCmd, removeCmd)
	
	// Verify subcommands are registered
	commands := cmd.Commands()
	assert.Len(t, commands, 3, "should have 3 subcommands")
	
	commandNames := make([]string, len(commands))
	for i, c := range commands {
		commandNames[i] = c.Use
	}
	assert.Contains(t, commandNames, "add")
	assert.Contains(t, commandNames, "list")
	assert.Contains(t, commandNames, "remove")
}

func TestGlobalRepoFlag(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantRepo   string
		beforeTest func()
		afterTest  func()
	}{
		{
			name:     "short flag",
			args:     []string{"-R", "owner/repo", "--help"},
			wantRepo: "owner/repo",
		},
		{
			name:     "long flag",
			args:     []string{"--repo", "owner/repo", "--help"},
			wantRepo: "owner/repo",
		},
		{
			name:     "no flag",
			args:     []string{"--help"},
			wantRepo: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeTest != nil {
				tt.beforeTest()
			}
			if tt.afterTest != nil {
				defer tt.afterTest()
			}
			
			// Reset repoFlag to default
			repoFlag = ""
			
			cmd := createTestRootCmd()
			cmd.SetArgs(tt.args)
			
			// Capture output to prevent it from showing during tests
			outBuf := &bytes.Buffer{}
			cmd.SetOut(outBuf)
			
			err := cmd.Execute()
			assert.NoError(t, err)
			
			assert.Equal(t, tt.wantRepo, repoFlag, "repoFlag should be set correctly")
		})
	}
}

func TestRootCmdUsageFormatting(t *testing.T) {
	cmd := createTestRootCmd()
	
	// Capture help output
	outBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)
	cmd.SetArgs([]string{"--help"})
	
	err := cmd.Execute()
	assert.NoError(t, err)
	
	output := outBuf.String()
	
	// Test help formatting
	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "Flags:")
	assert.Contains(t, output, "--repo")
	assert.Contains(t, output, "-R")
	assert.Contains(t, output, "--help")
	assert.Contains(t, output, "--version")
}

// createTestRootCmd creates a fresh root command for testing to avoid shared state
func createTestRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gh-issue-dependency",
		Short: "Manage issue dependencies in GitHub repositories",
		Long: `A GitHub CLI extension for managing issue dependencies using GitHub's native dependency API.

This tool allows you to:
- View dependency relationships between issues
- Add dependencies between issues  
- Remove existing dependencies
- Manage cross-repository issue dependencies

Examples:
  gh issue-dependency list
  gh issue-dependency add 123 456
  gh issue-dependency remove 123 456
  gh issue-dependency list --repo owner/repo`,
		Version: Version,
		Run: func(cmd *cobra.Command, args []string) {
			// If no subcommand is specified, show help
			cmd.Help()
		},
	}
	
	// Add persistent flags
	cmd.PersistentFlags().StringVarP(&repoFlag, "repo", "R", "", "Select another repository using the [HOST/]OWNER/REPO format")
	
	// Set version template
	cmd.SetVersionTemplate("gh-issue-dependency version {{.Version}}\n")
	
	return cmd
}

func TestInit(t *testing.T) {
	// Test that init function has been called (flags should be set up)
	assert.NotNil(t, rootCmd.PersistentFlags().Lookup("repo"), "repo flag should be initialized")
}

// Test with environment variables
func TestRootCmdWithEnvVars(t *testing.T) {
	// Save original env
	originalGHToken := os.Getenv("GH_TOKEN") 
	originalGHRepo := os.Getenv("GH_REPO")
	
	defer func() {
		if originalGHToken != "" {
			os.Setenv("GH_TOKEN", originalGHToken)
		} else {
			os.Unsetenv("GH_TOKEN")
		}
		if originalGHRepo != "" {
			os.Setenv("GH_REPO", originalGHRepo)
		} else {
			os.Unsetenv("GH_REPO")
		}
	}()
	
	// Set test environment
	os.Setenv("GH_TOKEN", "test-token")
	os.Setenv("GH_REPO", "test-owner/test-repo")
	
	cmd := createTestRootCmd()
	cmd.SetArgs([]string{"--help"})
	
	// Capture output
	outBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)
	
	err := cmd.Execute()
	assert.NoError(t, err)
	
	// Command should still work with env vars present
	assert.Contains(t, outBuf.String(), "gh-issue-dependency")
}

func TestVersionConstant(t *testing.T) {
	// Test that Version is accessible and has expected default
	assert.NotEmpty(t, Version, "Version should not be empty")
	// In tests, Version should be "dev" unless overridden at build time
	if strings.Contains(Version, "dev") || strings.Contains(Version, "test") {
		t.Logf("Version is set to development/test value: %s", Version)
	}
}