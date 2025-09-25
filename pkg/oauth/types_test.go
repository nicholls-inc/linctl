package oauth

import (
	"encoding/json"
	"testing"
)

func TestTokenResponse_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		token    TokenResponse
		expected string
	}{
		{
			name: "complete token response",
			token: TokenResponse{
				AccessToken: "test-access-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
				Scope:       "read write issues:create",
			},
			expected: `{"access_token":"test-access-token","token_type":"Bearer","expires_in":3600,"scope":"read write issues:create"}`,
		},
		{
			name: "minimal token response",
			token: TokenResponse{
				AccessToken: "minimal-token",
				TokenType:   "Bearer",
			},
			expected: `{"access_token":"minimal-token","token_type":"Bearer","expires_in":0,"scope":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			jsonData, err := json.Marshal(tt.token)
			if err != nil {
				t.Fatalf("Failed to marshal token: %v", err)
			}

			if string(jsonData) != tt.expected {
				t.Errorf("Expected JSON %s, got %s", tt.expected, string(jsonData))
			}

			// Test unmarshaling
			var unmarshaled TokenResponse
			err = json.Unmarshal(jsonData, &unmarshaled)
			if err != nil {
				t.Fatalf("Failed to unmarshal token: %v", err)
			}

			if unmarshaled.AccessToken != tt.token.AccessToken {
				t.Errorf("Expected AccessToken %s, got %s", tt.token.AccessToken, unmarshaled.AccessToken)
			}
			if unmarshaled.TokenType != tt.token.TokenType {
				t.Errorf("Expected TokenType %s, got %s", tt.token.TokenType, unmarshaled.TokenType)
			}
			if unmarshaled.ExpiresIn != tt.token.ExpiresIn {
				t.Errorf("Expected ExpiresIn %d, got %d", tt.token.ExpiresIn, unmarshaled.ExpiresIn)
			}
			if unmarshaled.Scope != tt.token.Scope {
				t.Errorf("Expected Scope %s, got %s", tt.token.Scope, unmarshaled.Scope)
			}
		})
	}
}

func TestTokenResponse_JSONDeserialization(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected TokenResponse
		hasError bool
	}{
		{
			name:     "valid complete response",
			jsonData: `{"access_token":"test-token","token_type":"Bearer","expires_in":7200,"scope":"read write"}`,
			expected: TokenResponse{
				AccessToken: "test-token",
				TokenType:   "Bearer",
				ExpiresIn:   7200,
				Scope:       "read write",
			},
			hasError: false,
		},
		{
			name:     "missing optional fields",
			jsonData: `{"access_token":"test-token"}`,
			expected: TokenResponse{
				AccessToken: "test-token",
				TokenType:   "",
				ExpiresIn:   0,
				Scope:       "",
			},
			hasError: false,
		},
		{
			name:     "invalid JSON",
			jsonData: `{"access_token":"test-token"`,
			expected: TokenResponse{},
			hasError: true,
		},
		{
			name:     "wrong field types",
			jsonData: `{"access_token":123,"token_type":"Bearer"}`,
			expected: TokenResponse{},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var token TokenResponse
			err := json.Unmarshal([]byte(tt.jsonData), &token)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if token.AccessToken != tt.expected.AccessToken {
				t.Errorf("Expected AccessToken %s, got %s", tt.expected.AccessToken, token.AccessToken)
			}
			if token.TokenType != tt.expected.TokenType {
				t.Errorf("Expected TokenType %s, got %s", tt.expected.TokenType, token.TokenType)
			}
			if token.ExpiresIn != tt.expected.ExpiresIn {
				t.Errorf("Expected ExpiresIn %d, got %d", tt.expected.ExpiresIn, token.ExpiresIn)
			}
			if token.Scope != tt.expected.Scope {
				t.Errorf("Expected Scope %s, got %s", tt.expected.Scope, token.Scope)
			}
		})
	}
}
