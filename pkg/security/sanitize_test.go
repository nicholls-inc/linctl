package security

import (
	"strings"
	"testing"
)

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal text",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "text with null bytes",
			input:    "Hello\x00world",
			expected: "Helloworld",
		},
		{
			name:     "text with control characters",
			input:    "Hello\x01\x02world",
			expected: "Helloworld",
		},
		{
			name:     "text with allowed whitespace",
			input:    "Hello\n\r\t world",
			expected: "Hello world", // Excessive whitespace gets collapsed
		},
		{
			name:     "text with excessive whitespace",
			input:    "  Hello    world  ",
			expected: "Hello world",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   \t\n  ",
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := SanitizeInput(test.input)
			if result != test.expected {
				t.Errorf("SanitizeInput(%q) = %q, expected %q", test.input, result, test.expected)
			}
		})
	}
}

func TestValidateIssueID(t *testing.T) {
	tests := []struct {
		name      string
		issueID   string
		expectErr bool
	}{
		{
			name:      "valid issue ID",
			issueID:   "ENG-123",
			expectErr: false,
		},
		{
			name:      "valid issue ID with longer team",
			issueID:   "DESIGN-456",
			expectErr: false,
		},
		{
			name:      "valid issue ID with numbers in team",
			issueID:   "TEAM2-789",
			expectErr: false,
		},
		{
			name:      "empty issue ID",
			issueID:   "",
			expectErr: true,
		},
		{
			name:      "lowercase team",
			issueID:   "eng-123",
			expectErr: true,
		},
		{
			name:      "no dash",
			issueID:   "ENG123",
			expectErr: true,
		},
		{
			name:      "no number",
			issueID:   "ENG-",
			expectErr: true,
		},
		{
			name:      "invalid characters",
			issueID:   "ENG-123!",
			expectErr: true,
		},
		{
			name:      "too long",
			issueID:   "VERYLONGTEAMNAME-123456789",
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateIssueID(test.issueID)
			if test.expectErr && err == nil {
				t.Errorf("ValidateIssueID(%q) expected error but got none", test.issueID)
			}
			if !test.expectErr && err != nil {
				t.Errorf("ValidateIssueID(%q) expected no error but got: %v", test.issueID, err)
			}
		})
	}
}

func TestValidateTeamKey(t *testing.T) {
	tests := []struct {
		name      string
		teamKey   string
		expectErr bool
	}{
		{
			name:      "valid team key",
			teamKey:   "ENG",
			expectErr: false,
		},
		{
			name:      "valid team key with numbers",
			teamKey:   "TEAM2",
			expectErr: false,
		},
		{
			name:      "valid longer team key",
			teamKey:   "DESIGN",
			expectErr: false,
		},
		{
			name:      "empty team key",
			teamKey:   "",
			expectErr: true,
		},
		{
			name:      "lowercase",
			teamKey:   "eng",
			expectErr: true,
		},
		{
			name:      "starts with number",
			teamKey:   "2ENG",
			expectErr: true,
		},
		{
			name:      "too long",
			teamKey:   "VERYLONGTEAM",
			expectErr: true,
		},
		{
			name:      "single character",
			teamKey:   "E",
			expectErr: true,
		},
		{
			name:      "invalid characters",
			teamKey:   "ENG-",
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateTeamKey(test.teamKey)
			if test.expectErr && err == nil {
				t.Errorf("ValidateTeamKey(%q) expected error but got none", test.teamKey)
			}
			if !test.expectErr && err != nil {
				t.Errorf("ValidateTeamKey(%q) expected no error but got: %v", test.teamKey, err)
			}
		})
	}
}

func TestValidateTitle(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		expectErr bool
	}{
		{
			name:      "valid title",
			title:     "Fix authentication bug",
			expectErr: false,
		},
		{
			name:      "empty title",
			title:     "",
			expectErr: true,
		},
		{
			name:      "too short",
			title:     "Hi",
			expectErr: true,
		},
		{
			name:      "too long",
			title:     strings.Repeat("a", 256),
			expectErr: true,
		},
		{
			name:      "only whitespace",
			title:     "   \t\n  ",
			expectErr: true,
		},
		{
			name:      "minimum valid length",
			title:     "Fix",
			expectErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateTitle(test.title)
			if test.expectErr && err == nil {
				t.Errorf("ValidateTitle(%q) expected error but got none", test.title)
			}
			if !test.expectErr && err != nil {
				t.Errorf("ValidateTitle(%q) expected no error but got: %v", test.title, err)
			}
		})
	}
}

func TestValidateDescription(t *testing.T) {
	tests := []struct {
		name        string
		description string
		expectErr   bool
	}{
		{
			name:        "valid description",
			description: "This is a detailed description of the issue.",
			expectErr:   false,
		},
		{
			name:        "empty description",
			description: "",
			expectErr:   false, // Description is optional
		},
		{
			name:        "very long description",
			description: strings.Repeat("a", 50001),
			expectErr:   true,
		},
		{
			name:        "maximum valid length",
			description: strings.Repeat("a", 50000),
			expectErr:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateDescription(test.description)
			if test.expectErr && err == nil {
				t.Errorf("ValidateDescription expected error but got none")
			}
			if !test.expectErr && err != nil {
				t.Errorf("ValidateDescription expected no error but got: %v", err)
			}
		})
	}
}

func TestValidateActorName(t *testing.T) {
	tests := []struct {
		name      string
		actor     string
		expectErr bool
	}{
		{
			name:      "valid actor name",
			actor:     "AI Agent",
			expectErr: false,
		},
		{
			name:      "empty actor name",
			actor:     "",
			expectErr: false, // Actor is optional
		},
		{
			name:      "too long",
			actor:     strings.Repeat("a", 101),
			expectErr: true,
		},
		{
			name:      "only whitespace",
			actor:     "   \t  ",
			expectErr: true,
		},
		{
			name:      "maximum valid length",
			actor:     strings.Repeat("a", 100),
			expectErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateActorName(test.actor)
			if test.expectErr && err == nil {
				t.Errorf("ValidateActorName(%q) expected error but got none", test.actor)
			}
			if !test.expectErr && err != nil {
				t.Errorf("ValidateActorName(%q) expected no error but got: %v", test.actor, err)
			}
		})
	}
}

func TestValidateAvatarURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		expectErr bool
	}{
		{
			name:      "valid HTTPS URL",
			url:       "https://example.com/avatar.png",
			expectErr: false,
		},
		{
			name:      "empty URL",
			url:       "",
			expectErr: false, // Avatar URL is optional
		},
		{
			name:      "HTTP URL",
			url:       "http://example.com/avatar.png",
			expectErr: true, // Must be HTTPS
		},
		{
			name:      "invalid URL format",
			url:       "not-a-url",
			expectErr: true,
		},
		{
			name:      "too long URL",
			url:       "https://example.com/" + strings.Repeat("a", 2048),
			expectErr: true,
		},
		{
			name:      "URL with query parameters",
			url:       "https://example.com/avatar.png?size=256&format=png",
			expectErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateAvatarURL(test.url)
			if test.expectErr && err == nil {
				t.Errorf("ValidateAvatarURL(%q) expected error but got none", test.url)
			}
			if !test.expectErr && err != nil {
				t.Errorf("ValidateAvatarURL(%q) expected no error but got: %v", test.url, err)
			}
		})
	}
}

func TestValidatePriority(t *testing.T) {
	tests := []struct {
		name      string
		priority  int
		expectErr bool
	}{
		{
			name:      "valid priority 0",
			priority:  0,
			expectErr: false,
		},
		{
			name:      "valid priority 4",
			priority:  4,
			expectErr: false,
		},
		{
			name:      "valid priority 2",
			priority:  2,
			expectErr: false,
		},
		{
			name:      "invalid negative priority",
			priority:  -1,
			expectErr: true,
		},
		{
			name:      "invalid high priority",
			priority:  5,
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidatePriority(test.priority)
			if test.expectErr && err == nil {
				t.Errorf("ValidatePriority(%d) expected error but got none", test.priority)
			}
			if !test.expectErr && err != nil {
				t.Errorf("ValidatePriority(%d) expected no error but got: %v", test.priority, err)
			}
		})
	}
}

func TestSanitizeAndValidateAll(t *testing.T) {
	fields := map[string]interface{}{
		"issue_id":    "ENG-123",
		"team_key":    "ENG",
		"title":       "Fix bug",
		"description": "This is a description",
		"actor":       "AI Agent",
		"avatar_url":  "https://example.com/avatar.png",
		"priority":    2,
		"other_field": "some value",
	}

	sanitized, errors := SanitizeAndValidateAll(fields)

	if len(errors) > 0 {
		t.Errorf("Expected no validation errors, got: %v", errors)
	}

	if len(sanitized) != len(fields) {
		t.Errorf("Expected %d sanitized fields, got %d", len(fields), len(sanitized))
	}

	// Test with invalid data
	invalidFields := map[string]interface{}{
		"issue_id": "invalid-id",
		"team_key": "invalid",
		"title":    "", // Empty title
		"priority": 10, // Invalid priority
	}

	_, errors = SanitizeAndValidateAll(invalidFields)

	if len(errors) == 0 {
		t.Error("Expected validation errors for invalid fields")
	}

	// Should have errors for issue_id, team_key, title, and priority
	if len(errors) < 4 {
		t.Errorf("Expected at least 4 validation errors, got %d", len(errors))
	}
}

func TestIsValidInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "normal text",
			input:    "Hello world",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "text with null byte",
			input:    "Hello\x00world",
			expected: false,
		},
		{
			name:     "text with some control chars",
			input:    "Hello\n\r\tworld",
			expected: true,
		},
		{
			name:     "text with too many control chars",
			input:    "Hello\x01\x02\x03\x04\x05\x06world",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsValidInput(test.input)
			if result != test.expected {
				t.Errorf("IsValidInput(%q) = %v, expected %v", test.input, result, test.expected)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := ValidationError{
		Field:   "test_field",
		Value:   "test_value",
		Message: "test message",
	}

	expected := "validation error for field 'test_field': test message"
	if err.Error() != expected {
		t.Errorf("ValidationError.Error() = %q, expected %q", err.Error(), expected)
	}
}
