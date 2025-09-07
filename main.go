// Package main provides the entry point for the gh-issue-dependency GitHub CLI extension.
//
// This extension allows users to manage issue dependencies in GitHub repositories
// using GitHub's native dependency API. It provides commands to list, add, and
// remove dependency relationships between issues, both within the same repository
// and across different repositories.
package main

import (
	"os"

	"github.com/torynet/gh-issue-dependency/cmd"
)

// main is the application entry point that executes the root command and
// exits with the appropriate status code based on command execution results.
//
// Exit codes follow GitHub CLI conventions:
//   - 0: Success
//   - 1: General error
//   - 2: Invalid input/validation error
//   - 3: Permission denied
//   - 4: Authentication required
func main() {
	code := cmd.Execute()
	os.Exit(code)
}
