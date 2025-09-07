package pkg

import (
	"os/exec"
	"testing"
)

func TestParseRepoFlag(t *testing.T) {
	tests := []struct {
		name     string
		repoFlag string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "valid OWNER/REPO format",
			repoFlag:  "octocat/Hello-World",
			wantOwner: "octocat",
			wantRepo:  "Hello-World",
			wantErr:   false,
		},
		{
			name:      "valid GitHub URL",
			repoFlag:  "https://github.com/octocat/Hello-World",
			wantOwner: "octocat",
			wantRepo:  "Hello-World",
			wantErr:   false,
		},
		{
			name:      "valid HOST/OWNER/REPO format",
			repoFlag:  "github.com/octocat/Hello-World",
			wantOwner: "octocat",
			wantRepo:  "Hello-World",
			wantErr:   false,
		},
		{
			name:     "empty string",
			repoFlag: "",
			wantErr:  true,
		},
		{
			name:     "invalid format - single word",
			repoFlag: "invalid",
			wantErr:  true,
		},
		{
			name:     "invalid format - too many slashes",
			repoFlag: "host/owner/repo/extra",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := ParseRepoFlag(tt.repoFlag)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRepoFlag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if owner != tt.wantOwner {
					t.Errorf("ParseRepoFlag() owner = %v, want %v", owner, tt.wantOwner)
				}
				if repo != tt.wantRepo {
					t.Errorf("ParseRepoFlag() repo = %v, want %v", repo, tt.wantRepo)
				}
			}
		})
	}
}

func TestParseIssueURL(t *testing.T) {
	tests := []struct {
		name            string
		url             string
		wantOwner       string
		wantRepo        string
		wantIssueNumber int
		wantErr         bool
	}{
		{
			name:            "valid issue URL",
			url:             "https://github.com/octocat/Hello-World/issues/123",
			wantOwner:       "octocat",
			wantRepo:        "Hello-World",
			wantIssueNumber: 123,
			wantErr:         false,
		},
		{
			name:            "valid issue URL with query params",
			url:             "https://github.com/octocat/Hello-World/issues/456?tab=overview",
			wantOwner:       "octocat",
			wantRepo:        "Hello-World",
			wantIssueNumber: 456,
			wantErr:         false,
		},
		{
			name:            "valid issue URL with fragment",
			url:             "https://github.com/octocat/Hello-World/issues/789#issuecomment-123",
			wantOwner:       "octocat",
			wantRepo:        "Hello-World",
			wantIssueNumber: 789,
			wantErr:         false,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "non-GitHub URL",
			url:     "https://example.com/issues/123",
			wantErr: true,
		},
		{
			name:    "GitHub URL but not issues",
			url:     "https://github.com/octocat/Hello-World",
			wantErr: true,
		},
		{
			name:    "GitHub URL with invalid issue number",
			url:     "https://github.com/octocat/Hello-World/issues/abc",
			wantErr: true,
		},
		{
			name:    "GitHub URL with zero issue number",
			url:     "https://github.com/octocat/Hello-World/issues/0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, issueNumber, err := ParseIssueURL(tt.url)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseIssueURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if owner != tt.wantOwner {
					t.Errorf("ParseIssueURL() owner = %v, want %v", owner, tt.wantOwner)
				}
				if repo != tt.wantRepo {
					t.Errorf("ParseIssueURL() repo = %v, want %v", repo, tt.wantRepo)
				}
				if issueNumber != tt.wantIssueNumber {
					t.Errorf("ParseIssueURL() issueNumber = %v, want %v", issueNumber, tt.wantIssueNumber)
				}
			}
		})
	}
}

func TestIsGhNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "command not found",
			err:      &exec.Error{Name: "gh", Err: exec.ErrNotFound},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGhNotFound(tt.err)
			if result != tt.expected {
				t.Errorf("isGhNotFound() = %v, want %v", result, tt.expected)
			}
		})
	}
}