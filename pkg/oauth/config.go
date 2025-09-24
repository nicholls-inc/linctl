package oauth

import (
	"fmt"
	"os"
	"strings"
)

// Config represents OAuth configuration
type Config struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	BaseURL      string   `json:"base_url"`
	Scopes       []string `json:"scopes"`
}

// ActorConfig represents default actor configuration
type ActorConfig struct {
	DefaultActor     string `json:"default_actor"`
	DefaultAvatarURL string `json:"default_avatar_url"`
}

// DefaultScopes returns the default OAuth scopes for Linear
func DefaultScopes() []string {
	return []string{"read", "write", "issues:create", "comments:create"}
}

// LoadFromEnvironment loads OAuth configuration from environment variables
func LoadFromEnvironment() (*Config, error) {
	clientID := os.Getenv("LINEAR_CLIENT_ID")
	clientSecret := os.Getenv("LINEAR_CLIENT_SECRET")
	baseURL := os.Getenv("LINEAR_BASE_URL")
	scopesEnv := os.Getenv("LINEAR_SCOPES")

	// Set defaults
	if baseURL == "" {
		baseURL = "https://api.linear.app"
	}

	var scopes []string
	if scopesEnv != "" {
		// Split scopes by comma and trim whitespace
		for _, scope := range strings.Split(scopesEnv, ",") {
			scope = strings.TrimSpace(scope)
			if scope != "" {
				scopes = append(scopes, scope)
			}
		}
	} else {
		scopes = DefaultScopes()
	}

	config := &Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		BaseURL:      baseURL,
		Scopes:       scopes,
	}

	return config, nil
}

// LoadActorFromEnvironment loads actor configuration from environment variables
func LoadActorFromEnvironment() *ActorConfig {
	return &ActorConfig{
		DefaultActor:     os.Getenv("LINEAR_DEFAULT_ACTOR"),
		DefaultAvatarURL: os.Getenv("LINEAR_DEFAULT_AVATAR_URL"),
	}
}

// IsConfigured returns true if actor configuration is available
func (ac *ActorConfig) IsConfigured() bool {
	return ac != nil && (ac.DefaultActor != "" || ac.DefaultAvatarURL != "")
}

// GetActor returns the actor name, using the provided value or falling back to default
func (ac *ActorConfig) GetActor(provided string) string {
	if provided != "" {
		return provided
	}
	if ac != nil {
		return ac.DefaultActor
	}
	return ""
}

// GetAvatarURL returns the avatar URL, using the provided value or falling back to default
func (ac *ActorConfig) GetAvatarURL(provided string) string {
	if provided != "" {
		return provided
	}
	if ac != nil {
		return ac.DefaultAvatarURL
	}
	return ""
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if c.ClientID == "" {
		return fmt.Errorf("client ID is required (set LINEAR_CLIENT_ID environment variable)")
	}

	if c.ClientSecret == "" {
		return fmt.Errorf("client secret is required (set LINEAR_CLIENT_SECRET environment variable)")
	}

	if c.BaseURL == "" {
		return fmt.Errorf("base URL cannot be empty")
	}

	if len(c.Scopes) == 0 {
		return fmt.Errorf("at least one scope is required")
	}

	// Validate URL format (basic check)
	if !strings.HasPrefix(c.BaseURL, "http://") && !strings.HasPrefix(c.BaseURL, "https://") {
		return fmt.Errorf("base URL must start with http:// or https://")
	}

	return nil
}

// IsComplete checks if all required fields are present
func (c *Config) IsComplete() bool {
	return c != nil && c.ClientID != "" && c.ClientSecret != ""
}

// GetScopesString returns scopes as a space-separated string
func (c *Config) GetScopesString() string {
	if c == nil || len(c.Scopes) == 0 {
		return strings.Join(DefaultScopes(), " ")
	}
	return strings.Join(c.Scopes, " ")
}

// HasScope checks if a specific scope is included
func (c *Config) HasScope(scope string) bool {
	if c == nil {
		return false
	}
	
	for _, s := range c.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// GetEnvironmentStatus returns information about environment variable configuration
func GetEnvironmentStatus() map[string]interface{} {
	status := map[string]interface{}{
		"LINEAR_CLIENT_ID":         os.Getenv("LINEAR_CLIENT_ID") != "",
		"LINEAR_CLIENT_SECRET":     os.Getenv("LINEAR_CLIENT_SECRET") != "",
		"LINEAR_BASE_URL":          os.Getenv("LINEAR_BASE_URL"),
		"LINEAR_SCOPES":            os.Getenv("LINEAR_SCOPES"),
		"LINEAR_DEFAULT_ACTOR":     os.Getenv("LINEAR_DEFAULT_ACTOR"),
		"LINEAR_DEFAULT_AVATAR_URL": os.Getenv("LINEAR_DEFAULT_AVATAR_URL"),
	}

	// Don't expose actual values for security
	if status["LINEAR_CLIENT_ID"].(bool) {
		status["LINEAR_CLIENT_ID"] = "set"
	} else {
		status["LINEAR_CLIENT_ID"] = "not set"
	}

	if status["LINEAR_CLIENT_SECRET"].(bool) {
		status["LINEAR_CLIENT_SECRET"] = "set"
	} else {
		status["LINEAR_CLIENT_SECRET"] = "not set"
	}

	if status["LINEAR_BASE_URL"].(string) == "" {
		status["LINEAR_BASE_URL"] = "not set (using default)"
	}

	if status["LINEAR_SCOPES"].(string) == "" {
		status["LINEAR_SCOPES"] = "not set (using defaults)"
	}

	if status["LINEAR_DEFAULT_ACTOR"].(string) == "" {
		status["LINEAR_DEFAULT_ACTOR"] = "not set"
	}

	if status["LINEAR_DEFAULT_AVATAR_URL"].(string) == "" {
		status["LINEAR_DEFAULT_AVATAR_URL"] = "not set"
	}

	return status
}