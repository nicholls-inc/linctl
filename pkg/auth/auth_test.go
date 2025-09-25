package auth

import (
	"encoding/json"
	"testing"
)

func TestAuthConfig_JSONSerialization_Minimal(t *testing.T) {
	tests := []struct {
		name     string
		config   AuthConfig
		expected string
	}{
		{
			name: "API key only",
			config: AuthConfig{
				APIKey: "test-api-key",
			},
			expected: `{"api_key":"test-api-key"}`,
		},
		{
			name:     "empty config",
			config:   AuthConfig{},
			expected: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.config)
			if err != nil {
				t.Fatalf("Failed to marshal config: %v", err)
			}

			if string(data) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(data))
			}

			// Test deserialization
			var config AuthConfig
			err = json.Unmarshal(data, &config)
			if err != nil {
				t.Fatalf("Failed to unmarshal config: %v", err)
			}

			if config.APIKey != tt.config.APIKey {
				t.Errorf("Expected APIKey %s, got %s", tt.config.APIKey, config.APIKey)
			}
		})
	}
}
