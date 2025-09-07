package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
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

// Global flags
var repoFlag string

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&repoFlag, "repo", "R", "", "Select another repository using the [HOST/]OWNER/REPO format")

	// Version template
	rootCmd.SetVersionTemplate("gh-issue-dependency version {{.Version}}\n")
}