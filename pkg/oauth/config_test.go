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