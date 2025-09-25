package integration

import (
	"os"
	"testing"

	"github.com/dorkitude/linctl/pkg/utils"
)

func TestActorParameterIntegration(t *testing.T) {
	// Save original environment
	originalActor := os.Getenv("LINEAR_DEFAULT_ACTOR")
	originalAvatarURL := os.Getenv("LINEAR_DEFAULT_AVATAR_URL")

	// Clean up after test
	defer func() {
		os.Setenv("LINEAR_DEFAULT_ACTOR", originalActor)
		os.Setenv("LINEAR_DEFAULT_AVATAR_URL", originalAvatarURL)
	}()

	tests := []struct {
		name           string
		envActor       string
		envAvatarURL   string
		flagActor      string
		flagAvatarURL  string
		expectedActor  string
		expectedAvatar string
	}{
		{
			name:           "no environment or flags",
			envActor:       "",
			envAvatarURL:   "",
			flagActor:      "",
			flagAvatarURL:  "",
			expectedActor:  "",
			expectedAvatar: "",
		},
		{
			name:           "environment variables only",
			envActor:       "Env Agent",
			envAvatarURL:   "https://env.com/avatar.png",
			flagActor:      "",
			flagAvatarURL:  "",
			expectedActor:  "Env Agent",
			expectedAvatar: "https://env.com/avatar.png",
		},
		{
			name:           "flags only",
			envActor:       "",
			envAvatarURL:   "",
			flagActor:      "Flag Agent",
			flagAvatarURL:  "https://flag.com/avatar.png",
			expectedActor:  "Flag Agent",
			expectedAvatar: "https://flag.com/avatar.png",
		},
		{
			name:           "flags override environment",
			envActor:       "Env Agent",
			envAvatarURL:   "https://env.com/avatar.png",
			flagActor:      "Flag Agent",
			flagAvatarURL:  "https://flag.com/avatar.png",
			expectedActor:  "Flag Agent",
			expectedAvatar: "https://flag.com/avatar.png",
		},
		{
			name:           "partial flag override - actor only",
			envActor:       "Env Agent",
			envAvatarURL:   "https://env.com/avatar.png",
			flagActor:      "Flag Agent",
			flagAvatarURL:  "",
			expectedActor:  "Flag Agent",
			expectedAvatar: "https://env.com/avatar.png",
		},
		{
			name:           "partial flag override - avatar only",
			envActor:       "Env Agent",
			envAvatarURL:   "https://env.com/avatar.png",
			flagActor:      "",
			flagAvatarURL:  "https://flag.com/avatar.png",
			expectedActor:  "Env Agent",
			expectedAvatar: "https://flag.com/avatar.png",
		},
		{
			name:           "empty flags don't override environment",
			envActor:       "Env Agent",
			envAvatarURL:   "https://env.com/avatar.png",
			flagActor:      "",
			flagAvatarURL:  "",
			expectedActor:  "Env Agent",
			expectedAvatar: "https://env.com/avatar.png",
		},
		{
			name:           "whitespace-only flags don't override environment",
			envActor:       "Env Agent",
			envAvatarURL:   "https://env.com/avatar.png",
			flagActor:      "   ",
			flagAvatarURL:  "   ",
			expectedActor:  "Env Agent",
			expectedAvatar: "https://env.com/avatar.png",
		},
		{
			name:           "complex actor names",
			envActor:       "",
			envAvatarURL:   "",
			flagActor:      "AI Agent (v2.0) - Production",
			flagAvatarURL:  "https://cdn.example.com/avatars/ai-agent-v2.png?size=64&format=png",
			expectedActor:  "AI Agent (v2.0) - Production",
			expectedAvatar: "https://cdn.example.com/avatars/ai-agent-v2.png?size=64&format=png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("LINEAR_DEFAULT_ACTOR", tt.envActor)
			os.Setenv("LINEAR_DEFAULT_AVATAR_URL", tt.envAvatarURL)

			// Call the actor resolution function
			actorParams := utils.ResolveActorParams(tt.flagActor, tt.flagAvatarURL)

			// Verify results
			actualActor := ""
			actualAvatar := ""
			if actorParams != nil {
				actualActor = actorParams.Actor
				actualAvatar = actorParams.AvatarURL
			}

			if actualActor != tt.expectedActor {
				t.Errorf("Expected actor '%s', got '%s'", tt.expectedActor, actualActor)
			}

			if actualAvatar != tt.expectedAvatar {
				t.Errorf("Expected avatar URL '%s', got '%s'", tt.expectedAvatar, actualAvatar)
			}
		})
	}
}

func TestActorParameterValidation(t *testing.T) {
	// Skip this test since ValidateActorParams function doesn't exist yet
	// This would be implemented when validation is needed
	t.Skip("ValidateActorParams function not implemented yet")
}

func TestActorParameterSerialization(t *testing.T) {
	// Skip this test since SerializeActorParams function doesn't exist yet
	// This would be implemented when serialization is needed
	t.Skip("SerializeActorParams function not implemented yet")
}

func TestActorParameterEndToEndFlow(t *testing.T) {
	// Save original environment
	originalActor := os.Getenv("LINEAR_DEFAULT_ACTOR")
	originalAvatarURL := os.Getenv("LINEAR_DEFAULT_AVATAR_URL")

	// Clean up after test
	defer func() {
		os.Setenv("LINEAR_DEFAULT_ACTOR", originalActor)
		os.Setenv("LINEAR_DEFAULT_AVATAR_URL", originalAvatarURL)
	}()

	// Test the complete flow from environment/flags to API parameters
	tests := []struct {
		name           string
		envActor       string
		envAvatarURL   string
		flagActor      string
		flagAvatarURL  string
		expectedParams map[string]interface{}
	}{
		{
			name:          "complete flow with flags",
			envActor:      "Env Agent",
			envAvatarURL:  "https://env.com/avatar.png",
			flagActor:     "Flag Agent",
			flagAvatarURL: "https://flag.com/avatar.png",
			expectedParams: map[string]interface{}{
				"actor":     "Flag Agent",
				"avatarUrl": "https://flag.com/avatar.png",
			},
		},
		{
			name:          "complete flow with environment only",
			envActor:      "Env Agent",
			envAvatarURL:  "https://env.com/avatar.png",
			flagActor:     "",
			flagAvatarURL: "",
			expectedParams: map[string]interface{}{
				"actor":     "Env Agent",
				"avatarUrl": "https://env.com/avatar.png",
			},
		},
		{
			name:           "complete flow with no actor info",
			envActor:       "",
			envAvatarURL:   "",
			flagActor:      "",
			flagAvatarURL:  "",
			expectedParams: map[string]interface{}{},
		},
		{
			name:          "complete flow with partial override",
			envActor:      "Env Agent",
			envAvatarURL:  "https://env.com/avatar.png",
			flagActor:     "Flag Agent",
			flagAvatarURL: "",
			expectedParams: map[string]interface{}{
				"actor":     "Flag Agent",
				"avatarUrl": "https://env.com/avatar.png",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment
			os.Setenv("LINEAR_DEFAULT_ACTOR", tt.envActor)
			os.Setenv("LINEAR_DEFAULT_AVATAR_URL", tt.envAvatarURL)

			// Resolve actor parameters
			actorParams := utils.ResolveActorParams(tt.flagActor, tt.flagAvatarURL)

			// Extract values for comparison
			actor := ""
			avatarURL := ""
			if actorParams != nil {
				actor = actorParams.Actor
				avatarURL = actorParams.AvatarURL
			}

			// Create expected API params manually for testing
			apiParams := make(map[string]interface{})
			if actor != "" {
				apiParams["actor"] = actor
			}
			if avatarURL != "" {
				apiParams["avatarUrl"] = avatarURL
			}

			// Verify final result matches expectations
			for key, expectedValue := range tt.expectedParams {
				actualValue, exists := apiParams[key]
				if !exists {
					t.Errorf("Expected key '%s' to be present in API params", key)
					continue
				}
				if actualValue != expectedValue {
					t.Errorf("Expected %s='%v', got '%v'", key, expectedValue, actualValue)
				}
			}

			// Check that no unexpected keys are present
			for key := range apiParams {
				if _, expected := tt.expectedParams[key]; !expected {
					t.Errorf("Unexpected key '%s' in API params with value '%v'", key, apiParams[key])
				}
			}

			// Verify the result has the correct number of keys
			if len(apiParams) != len(tt.expectedParams) {
				t.Errorf("Expected %d keys in API params, got %d", len(tt.expectedParams), len(apiParams))
			}
		})
	}
}

func TestActorParameterConcurrency(t *testing.T) {
	// Test that actor parameter resolution is thread-safe
	// This is important for CLI tools that might process multiple commands concurrently

	// Save original environment
	originalActor := os.Getenv("LINEAR_DEFAULT_ACTOR")
	originalAvatarURL := os.Getenv("LINEAR_DEFAULT_AVATAR_URL")

	// Clean up after test
	defer func() {
		os.Setenv("LINEAR_DEFAULT_ACTOR", originalActor)
		os.Setenv("LINEAR_DEFAULT_AVATAR_URL", originalAvatarURL)
	}()

	// Set base environment
	os.Setenv("LINEAR_DEFAULT_ACTOR", "Base Agent")
	os.Setenv("LINEAR_DEFAULT_AVATAR_URL", "https://base.com/avatar.png")

	// Run multiple goroutines that resolve actor parameters
	const numGoroutines = 10
	const numIterations = 100

	results := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			success := true
			for j := 0; j < numIterations; j++ {
				// Each goroutine uses different flag values
				flagActor := ""
				flagAvatarURL := ""
				expectedActor := "Base Agent"
				expectedAvatarURL := "https://base.com/avatar.png"

				if id%2 == 0 {
					flagActor = "Goroutine Agent"
					expectedActor = "Goroutine Agent"
				}

				if id%3 == 0 {
					flagAvatarURL = "https://goroutine.com/avatar.png"
					expectedAvatarURL = "https://goroutine.com/avatar.png"
				}

				actorParams := utils.ResolveActorParams(flagActor, flagAvatarURL)

				actor := ""
				avatarURL := ""
				if actorParams != nil {
					actor = actorParams.Actor
					avatarURL = actorParams.AvatarURL
				}

				if actor != expectedActor || avatarURL != expectedAvatarURL {
					success = false
					break
				}
			}
			results <- success
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		success := <-results
		if !success {
			t.Error("Concurrent actor parameter resolution failed")
		}
	}
}
