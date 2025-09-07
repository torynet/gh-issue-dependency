package pkg

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAppError(t *testing.T) {
	err := NewAppError(ErrorTypeValidation, "test message", errors.New("cause"))
	assert.Equal(t, ErrorTypeValidation, err.Type, "should have correct error type")
	assert.Equal(t, "test message", err.Message, "should have correct message")
	assert.Equal(t, "test message", err.Error(), "Error() should return message")
	assert.NotNil(t, err.Cause, "should have cause")
}

func TestAppErrorWithContext(t *testing.T) {
	err := NewAppError(ErrorTypeValidation, "test", nil).
		WithContext("key", "value").
		WithSuggestion("suggestion")

	if err.Context["key"] != "value" {
		t.Errorf("Expected context value 'value', got %v", err.Context["key"])
	}
	if len(err.Suggestions) != 1 || err.Suggestions[0] != "suggestion" {
		t.Errorf("Expected suggestion 'suggestion', got %v", err.Suggestions)
	}
}

func TestWrapAuthError(t *testing.T) {
	baseErr := errors.New("unauthorized")
	err := WrapAuthError(baseErr)

	if err.Type != ErrorTypeAuthentication {
		t.Errorf("Expected ErrorTypeAuthentication, got %v", err.Type)
	}
	if err.Cause != baseErr {
		t.Errorf("Expected wrapped error to be baseErr")
	}
	if len(err.Suggestions) == 0 {
		t.Error("Expected suggestions to be added")
	}
}

func TestParseIssueReference(t *testing.T) {
	tests := []struct {
		input    string
		wantRepo string
		wantNum  int
		wantErr  bool
	}{
		{"123", "", 123, false},
		{"owner/repo#456", "owner/repo", 456, false},
		{"https://github.com/owner/repo/issues/789", "owner/repo", 789, false},
		{"", "", 0, true},
		{"abc", "", 0, true},
		{"0", "", 0, true},
		{"-1", "", 0, true},
		{"owner#123", "", 0, true}, // Invalid repo format
	}

	for _, tt := range tests {
		repo, num, err := ParseIssueReference(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseIssueReference(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if err != nil {
			continue
		}
		if repo != tt.wantRepo {
			t.Errorf("ParseIssueReference(%q) repo = %v, want %v", tt.input, repo, tt.wantRepo)
		}
		if num != tt.wantNum {
			t.Errorf("ParseIssueReference(%q) num = %v, want %v", tt.input, num, tt.wantNum)
		}
	}
}

func TestValidateRepository(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"owner/repo", false},
		{"github/gh-cli", false},
		{"", true},
		{"owner", true},
		{"/repo", true},
		{"owner/", true},
		{"owner/repo/extra", true},
	}

	for _, tt := range tests {
		err := ValidateRepository(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateRepository(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
		}
	}
}

func TestWrapAPIError(t *testing.T) {
	tests := []struct {
		statusCode int
		wantType   ErrorType
	}{
		{http.StatusUnauthorized, ErrorTypeAuthentication},
		{http.StatusForbidden, ErrorTypePermission},
		{http.StatusNotFound, ErrorTypeAPI},
		{http.StatusTooManyRequests, ErrorTypeAPI},
		{http.StatusInternalServerError, ErrorTypeAPI},
		{http.StatusBadRequest, ErrorTypeAPI},
	}

	for _, tt := range tests {
		err := WrapAPIError(tt.statusCode, errors.New("test"))
		if err.Type != tt.wantType {
			t.Errorf("WrapAPIError(%d) type = %v, want %v", tt.statusCode, err.Type, tt.wantType)
		}
	}
}

func TestIsErrorType(t *testing.T) {
	authErr := WrapAuthError(errors.New("test"))
	validationErr := NewIssueNumberValidationError("abc")

	if !IsErrorType(authErr, ErrorTypeAuthentication) {
		t.Error("Expected authentication error to match ErrorTypeAuthentication")
	}
	if IsErrorType(authErr, ErrorTypeValidation) {
		t.Error("Expected authentication error not to match ErrorTypeValidation")
	}
	if !IsErrorType(validationErr, ErrorTypeValidation) {
		t.Error("Expected validation error to match ErrorTypeValidation")
	}

	// Test with non-AppError
	basicErr := errors.New("basic error")
	if IsErrorType(basicErr, ErrorTypeInternal) {
		t.Error("Expected basic error not to match any specific type")
	}
}

func TestGetErrorType(t *testing.T) {
	authErr := WrapAuthError(errors.New("test"))
	basicErr := errors.New("basic error")

	if GetErrorType(authErr) != ErrorTypeAuthentication {
		t.Error("Expected authentication error type")
	}
	if GetErrorType(basicErr) != ErrorTypeInternal {
		t.Error("Expected internal error type for basic error")
	}
}

func TestFormatUserError(t *testing.T) {
	err := NewAppError(ErrorTypeValidation, "Test error", nil).
		WithContext("field", "value").
		WithSuggestion("Try this")

	formatted := FormatUserError(err)
	
	// Should contain the main message
	if !contains(formatted, "Test error") {
		t.Error("Formatted error should contain main message")
	}
	
	// Should contain context
	if !contains(formatted, "field: value") {
		t.Error("Formatted error should contain context")
	}
	
	// Should contain suggestions
	if !contains(formatted, "Try this") {
		t.Error("Formatted error should contain suggestions")
	}
	
	// Test with basic error
	basicErr := errors.New("basic")
	basicFormatted := FormatUserError(basicErr)
	if !contains(basicFormatted, "basic") {
		t.Error("Should handle basic errors")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		stringContains(s, substr))))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}