package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestEnvironment manages environment variables for testing
type TestEnvironment struct {
	originalVars map[string]string
	tempDir      string
}

// NewTestEnvironment creates a new test environment with isolated config
func NewTestEnvironment(t *testing.T) *TestEnvironment {
	t.Helper()
	
	// Create temporary directory for test configs
	tempDir := t.TempDir()
	
	// Save original environment variables
	originalVars := map[string]string{
		"LINEAR_CLIENT_ID":         os.Getenv("LINEAR_CLIENT_ID"),
		"LINEAR_CLIENT_SECRET":     os.Getenv("LINEAR_CLIENT_SECRET"),
		"LINEAR_BASE_URL":          os.Getenv("LINEAR_BASE_URL"),
		"LINEAR_SCOPES":            os.Getenv("LINEAR_SCOPES"),
		"LINEAR_DEFAULT_ACTOR":     os.Getenv("LINEAR_DEFAULT_ACTOR"),
		"LINEAR_DEFAULT_AVATAR_URL": os.Getenv("LINEAR_DEFAULT_AVATAR_URL"),
	}
	
	return &TestEnvironment{
		originalVars: originalVars,
		tempDir:      tempDir,
	}
}

// ClearOAuthEnvironment removes all OAuth-related environment variables
func (te *TestEnvironment) ClearOAuthEnvironment() {
	os.Unsetenv("LINEAR_CLIENT_ID")
	os.Unsetenv("LINEAR_CLIENT_SECRET")
	os.Unsetenv("LINEAR_BASE_URL")
	os.Unsetenv("LINEAR_SCOPES")
	os.Unsetenv("LINEAR_DEFAULT_ACTOR")
	os.Unsetenv("LINEAR_DEFAULT_AVATAR_URL")
}

// SetOAuthEnvironment sets OAuth environment variables for testing
func (te *TestEnvironment) SetOAuthEnvironment(clientID, clientSecret string) {
	if clientID != "" {
		os.Setenv("LINEAR_CLIENT_ID", clientID)
	}
	if clientSecret != "" {
		os.Setenv("LINEAR_CLIENT_SECRET", clientSecret)
	}
}

// SetActorEnvironment sets actor environment variables for testing
func (te *TestEnvironment) SetActorEnvironment(actor, avatarURL string) {
	if actor != "" {
		os.Setenv("LINEAR_DEFAULT_ACTOR", actor)
	}
	if avatarURL != "" {
		os.Setenv("LINEAR_DEFAULT_AVATAR_URL", avatarURL)
	}
}

// GetTempConfigPath returns a temporary config path for testing
func (te *TestEnvironment) GetTempConfigPath() string {
	return filepath.Join(te.tempDir, ".linctl-auth.json")
}

// GetTempOAuthTokenPath returns a temporary OAuth token path for testing
func (te *TestEnvironment) GetTempOAuthTokenPath() string {
	return filepath.Join(te.tempDir, ".linctl-oauth-token.json")
}

// Cleanup restores original environment variables
func (te *TestEnvironment) Cleanup() {
	// Restore original environment variables
	for key, value := range te.originalVars {
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}
}

// MockAuthConfig creates a temporary auth config for testing
func (te *TestEnvironment) MockAuthConfig(config AuthConfig) error {
	// Save the config to the temp directory
	tempConfigPath := te.GetTempConfigPath()
	
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(tempConfigPath), 0700); err != nil {
		return err
	}
	
	// Write the config file directly
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(tempConfigPath, data, 0600)
}

// WithMockedConfigPath runs a function with a mocked config path
func (te *TestEnvironment) WithMockedConfigPath(fn func()) {
	// Override the config path to use temp directory
	originalGetConfigPath := getConfigPath
	getConfigPath = func() (string, error) {
		return te.GetTempConfigPath(), nil
	}
	
	// Restore original function when done
	defer func() {
		getConfigPath = originalGetConfigPath
	}()
	
	fn()
}

// WithIsolatedEnvironment runs a test function with an isolated environment
func WithIsolatedEnvironment(t *testing.T, fn func(*TestEnvironment)) {
	t.Helper()
	
	env := NewTestEnvironment(t)
	defer env.Cleanup()
	
	// Clear OAuth environment to prevent interference
	env.ClearOAuthEnvironment()
	
	fn(env)
}

// WithMockedOAuth runs a test function with mocked OAuth environment
func WithMockedOAuth(t *testing.T, clientID, clientSecret string, fn func(*TestEnvironment)) {
	t.Helper()
	
	WithIsolatedEnvironment(t, func(env *TestEnvironment) {
		env.SetOAuthEnvironment(clientID, clientSecret)
		fn(env)
	})
}