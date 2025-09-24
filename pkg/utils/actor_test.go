package utils

import (
	"os"
	"testing"
)

func TestResolveActorParams(t *testing.T) {
	// Save original environment
	originalActor := os.Getenv("LINEAR_DEFAULT_ACTOR")
	originalAvatarURL := os.Getenv("LINEAR_DEFAULT_AVATAR_URL")
	
	// Clean up after test
	defer func() {
		os.Setenv("LINEAR_DEFAULT_ACTOR", originalActor)
		os.Setenv("LINEAR_DEFAULT_AVATAR_URL", originalAvatarURL)
	}()

	tests := []struct {
		name             string
		envActor         string
		envAvatarURL     string
		providedActor    string
		providedAvatarURL string
		expectedActor    string
		expectedAvatarURL string
		hasActorInfo     bool
	}{
		{
			name:             "provided values take priority",
			envActor:         "Env Agent",
			envAvatarURL:     "https://env.com/avatar.png",
			providedActor:    "Provided Agent",
			providedAvatarURL: "https://provided.com/avatar.png",
			expectedActor:    "Provided Agent",
			expectedAvatarURL: "https://provided.com/avatar.png",
			hasActorInfo:     true,
		},
		{
			name:             "fallback to environment",
			envActor:         "Env Agent",
			envAvatarURL:     "https://env.com/avatar.png",
			providedActor:    "",
			providedAvatarURL: "",
			expectedActor:    "Env Agent",
			expectedAvatarURL: "https://env.com/avatar.png",
			hasActorInfo:     true,
		},
		{
			name:             "mixed provided and environment",
			envActor:         "Env Agent",
			envAvatarURL:     "https://env.com/avatar.png",
			providedActor:    "Provided Agent",
			providedAvatarURL: "",
			expectedActor:    "Provided Agent",
			expectedAvatarURL: "https://env.com/avatar.png",
			hasActorInfo:     true,
		},
		{
			name:             "no actor info available",
			envActor:         "",
			envAvatarURL:     "",
			providedActor:    "",
			providedAvatarURL: "",
			expectedActor:    "",
			expectedAvatarURL: "",
			hasActorInfo:     false,
		},
		{
			name:             "only actor provided",
			envActor:         "",
			envAvatarURL:     "",
			providedActor:    "Solo Agent",
			providedAvatarURL: "",
			expectedActor:    "Solo Agent",
			expectedAvatarURL: "",
			hasActorInfo:     true,
		},
		{
			name:             "only avatar URL provided",
			envActor:         "",
			envAvatarURL:     "",
			providedActor:    "",
			providedAvatarURL: "https://solo.com/avatar.png",
			expectedActor:    "",
			expectedAvatarURL: "https://solo.com/avatar.png",
			hasActorInfo:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("LINEAR_DEFAULT_ACTOR", tt.envActor)
			os.Setenv("LINEAR_DEFAULT_AVATAR_URL", tt.envAvatarURL)

			// Resolve actor parameters
			params := ResolveActorParams(tt.providedActor, tt.providedAvatarURL)

			// Test resolved values
			if params.Actor != tt.expectedActor {
				t.Errorf("Expected actor '%s', got '%s'", tt.expectedActor, params.Actor)
			}

			if params.AvatarURL != tt.expectedAvatarURL {
				t.Errorf("Expected avatar URL '%s', got '%s'", tt.expectedAvatarURL, params.AvatarURL)
			}

			// Test HasActorInfo
			if params.HasActorInfo() != tt.hasActorInfo {
				t.Errorf("Expected HasActorInfo() to return %v, got %v", tt.hasActorInfo, params.HasActorInfo())
			}

			// Test ToCreateAsUser
			createAsUser := params.ToCreateAsUser()
			if tt.expectedActor == "" {
				if createAsUser != nil {
					t.Errorf("Expected ToCreateAsUser() to return nil for empty actor, got %v", createAsUser)
				}
			} else {
				if createAsUser == nil {
					t.Error("Expected ToCreateAsUser() to return non-nil for non-empty actor")
				} else if *createAsUser != tt.expectedActor {
					t.Errorf("Expected ToCreateAsUser() to return '%s', got '%s'", tt.expectedActor, *createAsUser)
				}
			}

			// Test ToDisplayIconURL
			displayIconURL := params.ToDisplayIconURL()
			if tt.expectedAvatarURL == "" {
				if displayIconURL != nil {
					t.Errorf("Expected ToDisplayIconURL() to return nil for empty avatar URL, got %v", displayIconURL)
				}
			} else {
				if displayIconURL == nil {
					t.Error("Expected ToDisplayIconURL() to return non-nil for non-empty avatar URL")
				} else if *displayIconURL != tt.expectedAvatarURL {
					t.Errorf("Expected ToDisplayIconURL() to return '%s', got '%s'", tt.expectedAvatarURL, *displayIconURL)
				}
			}
		})
	}
}

func TestActorParamsNil(t *testing.T) {
	var params *ActorParams = nil

	// Test nil safety
	if params.HasActorInfo() {
		t.Error("Expected HasActorInfo() to return false for nil ActorParams")
	}

	if params.ToCreateAsUser() != nil {
		t.Error("Expected ToCreateAsUser() to return nil for nil ActorParams")
	}

	if params.ToDisplayIconURL() != nil {
		t.Error("Expected ToDisplayIconURL() to return nil for nil ActorParams")
	}
}