// Package cmd implements all CLI commands for the gh-issue-dependency extension.
//
// This package contains the command-line interface built with Cobra, including
// the root command and all subcommands (list, add, remove). Each command
// handles user input validation, interacts with the GitHub API through the
// pkg package, and provides structured error messages.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/torynet/gh-issue-dependency/pkg"
)

// Version contains the current version of the application.
// This is set during build time and displayed in version output.
var Version = "dev"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gh-issue-dependency",
	Short: "Manage GitHub issue dependencies",
	Long: `Manage issue dependencies in GitHub repositories using GitHub's native dependency API.

This extension helps you organize complex projects by creating dependency relationships
between issues, whether in the same repository or across different repositories.

USAGE
  gh issue-dependency <command>

CORE COMMANDS
  list     List issue dependencies and relationships
  add      Add dependency relationships between issues
  remove   Remove existing dependency relationships

FLAGS
  -R, --repo OWNER/REPO   Select repository using OWNER/REPO format

EXAMPLES
  # List all dependencies for issue #123
  gh issue-dependency list 123

  # Make issue #123 depend on issue #456  
  gh issue-dependency add 123 --blocked-by 456

  # Remove a dependency relationship
  gh issue-dependency remove 123 --blocked-by 456

  # Work with issues in a different repository
  gh issue-dependency list 123 --repo owner/other-repo

AUTHENTICATION
  This extension uses the same authentication as the GitHub CLI. Run 'gh auth status' 
  to check your authentication status. Use 'gh auth login' if you need to authenticate.

LEARN MORE
  Use 'gh issue-dependency <command> --help' for more information about a specific command.`,
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is specified, show help
		_ = cmd.Help()
	},
}

// Global flags accessible to all commands
var repoFlag string

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
//
// Returns an exit code following GitHub CLI conventions:
//   - 0: Success
//   - 1: General error
//   - 2: Invalid input/validation error
//   - 3: Permission denied
//   - 4: Authentication required
func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		// Use our structured error formatting for user-friendly messages
		fmt.Fprintf(os.Stderr, "%s\n", pkg.FormatUserError(err))

		// Return appropriate exit codes based on error type
		switch pkg.GetErrorType(err) {
		case pkg.ErrorTypeAuthentication:
			return 4 // Authentication required
		case pkg.ErrorTypePermission:
			return 3 // Permission denied
		case pkg.ErrorTypeValidation:
			return 2 // Invalid input
		default:
			return 1 // General error
		}
	}
	return 0
}

// init initializes the root command with global flags and configuration.
// This function is called automatically when the package is imported.
func init() {
	// Global flags available to all commands
	rootCmd.PersistentFlags().StringVarP(&repoFlag, "repo", "R", "", "Select another repository using the [HOST/]OWNER/REPO format")

	// Configure version output template to match GitHub CLI style
	rootCmd.SetVersionTemplate("gh-issue-dependency version {{.Version}}\n")
}
