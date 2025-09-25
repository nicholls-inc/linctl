package oauth

import (
	"os"
	"testing"
)

func TestConfigValidation(t *testing.T) {
	// Test valid config
	validConfig := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		BaseURL:      "https://api.linear.app",
		Scopes:       []string{"read", "write"},
	}

	if err := validConfig.Validate(); err != nil {
		t.Errorf("Valid config should pass validation: %v", err)
	}

	// Test missing client ID
	invalidConfig := &Config{
		ClientSecret: "test-client-secret",
		BaseURL:      "https://api.linear.app",
		Scopes:       []string{"read", "write"},
	}

	if err := invalidConfig.Validate(); err == nil {
		t.Error("Config without client ID should fail validation")
	}

	// Test missing client secret
	invalidConfig2 := &Config{
		ClientID: "test-client-id",
		BaseURL:  "https://api.linear.app",
		Scopes:   []string{"read", "write"},
	}

	if err := invalidConfig2.Validate(); err == nil {
		t.Error("Config without client secret should fail validation")
	}

	// Test invalid URL
	invalidConfig3 := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		BaseURL:      "invalid-url",
		Scopes:       []string{"read", "write"},
	}

	if err := invalidConfig3.Validate(); err == nil {
		t.Error("Config with invalid URL should fail validation")
	}

	// Test empty scopes
	invalidConfig4 := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		BaseURL:      "https://api.linear.app",
		Scopes:       []string{},
	}

	if err := invalidConfig4.Validate(); err == nil {
		t.Error("Config without scopes should fail validation")
	}
}

func TestConfigCompletion(t *testing.T) {
	// Test complete config
	completeConfig := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		BaseURL:      "https://api.linear.app",
		Scopes:       []string{"read", "write"},
	}

	if !completeConfig.IsComplete() {
		t.Error("Complete config should be reported as complete")
	}

	// Test incomplete config
	incompleteConfig := &Config{
		ClientID: "test-client-id",
		BaseURL:  "https://api.linear.app",
		Scopes:   []string{"read", "write"},
	}

	if incompleteConfig.IsComplete() {
		t.Error("Incomplete config should be reported as incomplete")
	}

	// Test nil config
	var nilConfig *Config
	if nilConfig.IsComplete() {
		t.Error("Nil config should be reported as incomplete")
	}
}

func TestScopeOperations(t *testing.T) {
	config := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		BaseURL:      "https://api.linear.app",
		Scopes:       []string{"read", "write", "issues:create"},
	}

	// Test scope string generation
	scopeString := config.GetScopesString()
	expected := "read write issues:create"
	if scopeString != expected {
		t.Errorf("Expected scope string '%s', got '%s'", expected, scopeString)
	}

	// Test scope checking
	if !config.HasScope("read") {
		t.Error("Config should have 'read' scope")
	}

	if !config.HasScope("issues:create") {
		t.Error("Config should have 'issues:create' scope")
	}

	if config.HasScope("admin") {
		t.Error("Config should not have 'admin' scope")
	}

	// Test nil config scope operations
	var nilConfig *Config
	if nilConfig.HasScope("read") {
		t.Error("Nil config should not have any scopes")
	}

	if nilConfig.GetScopesString() != "read write issues:create comments:create" {
		t.Error("Nil config should return default scopes")
	}
}

func TestEnvironmentLoading(t *testing.T) {
	// Save original environment
	originalClientID := os.Getenv("LINEAR_CLIENT_ID")
	originalClientSecret := os.Getenv("LINEAR_CLIENT_SECRET")
	originalBaseURL := os.Getenv("LINEAR_BASE_URL")
	originalScopes := os.Getenv("LINEAR_SCOPES")

	// Clean up after test
	defer func() {
		os.Setenv("LINEAR_CLIENT_ID", originalClientID)
		os.Setenv("LINEAR_CLIENT_SECRET", originalClientSecret)
		os.Setenv("LINEAR_BASE_URL", originalBaseURL)
		os.Setenv("LINEAR_SCOPES", originalScopes)
	}()

	// Test with environment variables set
	os.Setenv("LINEAR_CLIENT_ID", "env-client-id")
	os.Setenv("LINEAR_CLIENT_SECRET", "env-client-secret")
	os.Setenv("LINEAR_BASE_URL", "https://custom.linear.app")
	os.Setenv("LINEAR_SCOPES", "read,write,custom:scope")

	config, err := LoadFromEnvironment()
	if err != nil {
		t.Fatalf("Failed to load from environment: %v", err)
	}

	if config.ClientID != "env-client-id" {
		t.Errorf("Expected client ID 'env-client-id', got '%s'", config.ClientID)
	}

	if config.ClientSecret != "env-client-secret" {
		t.Errorf("Expected client secret 'env-client-secret', got '%s'", config.ClientSecret)
	}

	if config.BaseURL != "https://custom.linear.app" {
		t.Errorf("Expected base URL 'https://custom.linear.app', got '%s'", config.BaseURL)
	}

	expectedScopes := []string{"read", "write", "custom:scope"}
	if len(config.Scopes) != len(expectedScopes) {
		t.Errorf("Expected %d scopes, got %d", len(expectedScopes), len(config.Scopes))
	}

	for i, scope := range expectedScopes {
		if config.Scopes[i] != scope {
			t.Errorf("Expected scope '%s' at index %d, got '%s'", scope, i, config.Scopes[i])
		}
	}

	// Test with minimal environment (defaults)
	os.Unsetenv("LINEAR_BASE_URL")
	os.Unsetenv("LINEAR_SCOPES")

	config2, err := LoadFromEnvironment()
	if err != nil {
		t.Fatalf("Failed to load from environment with defaults: %v", err)
	}

	if config2.BaseURL != "https://api.linear.app" {
		t.Errorf("Expected default base URL, got '%s'", config2.BaseURL)
	}

	defaultScopes := DefaultScopes()
	if len(config2.Scopes) != len(defaultScopes) {
		t.Errorf("Expected default scopes count %d, got %d", len(defaultScopes), len(config2.Scopes))
	}
}

func TestDefaultScopes(t *testing.T) {
	scopes := DefaultScopes()
	expectedScopes := []string{"read", "write", "issues:create", "comments:create"}

	if len(scopes) != len(expectedScopes) {
		t.Errorf("Expected %d default scopes, got %d", len(expectedScopes), len(scopes))
	}

	for i, expected := range expectedScopes {
		if scopes[i] != expected {
			t.Errorf("Expected default scope '%s' at index %d, got '%s'", expected, i, scopes[i])
		}
	}
}

func TestActorConfig(t *testing.T) {
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
		expectedActor  string
		expectedAvatar string
		isConfigured   bool
	}{
		{
			name:           "both configured",
			envActor:       "AI Agent",
			envAvatarURL:   "https://example.com/agent.png",
			expectedActor:  "AI Agent",
			expectedAvatar: "https://example.com/agent.png",
			isConfigured:   true,
		},
		{
			name:           "only actor configured",
			envActor:       "AI Agent",
			envAvatarURL:   "",
			expectedActor:  "AI Agent",
			expectedAvatar: "",
			isConfigured:   true,
		},
		{
			name:           "only avatar configured",
			envActor:       "",
			envAvatarURL:   "https://example.com/agent.png",
			expectedActor:  "",
			expectedAvatar: "https://example.com/agent.png",
			isConfigured:   true,
		},
		{
			name:           "nothing configured",
			envActor:       "",
			envAvatarURL:   "",
			expectedActor:  "",
			expectedAvatar: "",
			isConfigured:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("LINEAR_DEFAULT_ACTOR", tt.envActor)
			os.Setenv("LINEAR_DEFAULT_AVATAR_URL", tt.envAvatarURL)

			// Load actor config
			config := LoadActorFromEnvironment()

			// Test basic fields
			if config.DefaultActor != tt.expectedActor {
				t.Errorf("Expected actor '%s', got '%s'", tt.expectedActor, config.DefaultActor)
			}

			if config.DefaultAvatarURL != tt.expectedAvatar {
				t.Errorf("Expected avatar URL '%s', got '%s'", tt.expectedAvatar, config.DefaultAvatarURL)
			}

			// Test IsConfigured
			if config.IsConfigured() != tt.isConfigured {
				t.Errorf("Expected IsConfigured() to return %v, got %v", tt.isConfigured, config.IsConfigured())
			}

			// Test GetActor with provided value
			providedActor := "Provided Actor"
			if result := config.GetActor(providedActor); result != providedActor {
				t.Errorf("Expected GetActor with provided value to return '%s', got '%s'", providedActor, result)
			}

			// Test GetActor with empty provided value
			if result := config.GetActor(""); result != tt.expectedActor {
				t.Errorf("Expected GetActor with empty provided value to return '%s', got '%s'", tt.expectedActor, result)
			}

			// Test GetAvatarURL with provided value
			providedURL := "https://provided.com/avatar.png"
			if result := config.GetAvatarURL(providedURL); result != providedURL {
				t.Errorf("Expected GetAvatarURL with provided value to return '%s', got '%s'", providedURL, result)
			}

			// Test GetAvatarURL with empty provided value
			if result := config.GetAvatarURL(""); result != tt.expectedAvatar {
				t.Errorf("Expected GetAvatarURL with empty provided value to return '%s', got '%s'", tt.expectedAvatar, result)
			}
		})
	}
}

func TestGetEnvironmentStatusWithActor(t *testing.T) {
	// Save original environment
	originalClientID := os.Getenv("LINEAR_CLIENT_ID")
	originalClientSecret := os.Getenv("LINEAR_CLIENT_SECRET")
	originalActor := os.Getenv("LINEAR_DEFAULT_ACTOR")
	originalAvatarURL := os.Getenv("LINEAR_DEFAULT_AVATAR_URL")

	// Clean up after test
	defer func() {
		os.Setenv("LINEAR_CLIENT_ID", originalClientID)
		os.Setenv("LINEAR_CLIENT_SECRET", originalClientSecret)
		os.Setenv("LINEAR_DEFAULT_ACTOR", originalActor)
		os.Setenv("LINEAR_DEFAULT_AVATAR_URL", originalAvatarURL)
	}()

	// Set test environment
	os.Setenv("LINEAR_CLIENT_ID", "test-client-id")
	os.Setenv("LINEAR_CLIENT_SECRET", "test-client-secret")
	os.Setenv("LINEAR_DEFAULT_ACTOR", "Test Agent")
	os.Setenv("LINEAR_DEFAULT_AVATAR_URL", "https://test.com/avatar.png")

	status := GetEnvironmentStatus()

	// Check that actor fields are included
	if _, ok := status["LINEAR_DEFAULT_ACTOR"]; !ok {
		t.Error("Expected LINEAR_DEFAULT_ACTOR in environment status")
	}

	if _, ok := status["LINEAR_DEFAULT_AVATAR_URL"]; !ok {
		t.Error("Expected LINEAR_DEFAULT_AVATAR_URL in environment status")
	}

	// Check values
	if status["LINEAR_DEFAULT_ACTOR"].(string) != "Test Agent" {
		t.Errorf("Expected LINEAR_DEFAULT_ACTOR to be 'Test Agent', got '%s'", status["LINEAR_DEFAULT_ACTOR"])
	}

	if status["LINEAR_DEFAULT_AVATAR_URL"].(string) != "https://test.com/avatar.png" {
		t.Errorf("Expected LINEAR_DEFAULT_AVATAR_URL to be 'https://test.com/avatar.png', got '%s'", status["LINEAR_DEFAULT_AVATAR_URL"])
	}
}
