package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dorkitude/linctl/pkg/auth"
	"github.com/dorkitude/linctl/pkg/oauth"
)

// AgentConfig represents configuration optimized for agent workflows
type AgentConfig struct {
	// Silent mode - suppress non-essential output
	Silent bool
	// JSON mode - force JSON output for all operations
	JSONMode bool
	// Timeout for operations (in seconds)
	Timeout int
	// Retry attempts for failed operations
	RetryAttempts int
	// Actor configuration
	DefaultActor     string
	DefaultAvatarURL string
}

// AgentResponse represents a standardized response for agent operations
type AgentResponse struct {
	Success   bool                   `json:"success"`
	Data      interface{}            `json:"data,omitempty"`
	Error     *AgentError            `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp string                 `json:"timestamp"`
}

// AgentError represents a structured error for agent consumption
type AgentError struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
	Retryable   bool                   `json:"retryable"`
}

// LoadAgentConfig loads configuration optimized for agent workflows
func LoadAgentConfig() *AgentConfig {
	config := &AgentConfig{
		Silent:           getBoolEnv("LINEAR_AGENT_SILENT", false),
		JSONMode:         getBoolEnv("LINEAR_AGENT_JSON", false),
		Timeout:          getIntEnv("LINEAR_AGENT_TIMEOUT", 30),
		RetryAttempts:    getIntEnv("LINEAR_AGENT_RETRY_ATTEMPTS", 3),
		DefaultActor:     os.Getenv("LINEAR_DEFAULT_ACTOR"),
		DefaultAvatarURL: os.Getenv("LINEAR_DEFAULT_AVATAR_URL"),
	}

	return config
}

// ValidateAgentEnvironment validates that the environment is properly configured for agent workflows
func ValidateAgentEnvironment() *AgentResponse {
	response := &AgentResponse{
		Success:   false,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Metadata:  make(map[string]interface{}),
	}

	// Check OAuth configuration
	if err := oauth.ValidateEnvironmentForAgent(); err != nil {
		response.Error = &AgentError{
			Code:    "OAUTH_CONFIG_ERROR",
			Message: err.Error(),
			Suggestions: []string{
				"Set LINEAR_CLIENT_ID environment variable",
				"Set LINEAR_CLIENT_SECRET environment variable",
				"Verify OAuth application is properly configured in Linear",
			},
			Retryable: false,
		}
		return response
	}

	// Check authentication status
	authStatus, err := auth.GetAuthStatus()
	if err != nil {
		response.Error = &AgentError{
			Code:    "AUTH_STATUS_ERROR",
			Message: fmt.Sprintf("Failed to get authentication status: %v", err),
			Suggestions: []string{
				"Check network connectivity",
				"Verify OAuth credentials are correct",
			},
			Retryable: true,
		}
		return response
	}

	if !authStatus.Authenticated {
		response.Error = &AgentError{
			Code:    "NOT_AUTHENTICATED",
			Message: "Not authenticated with Linear",
			Suggestions: []string{
				"Run authentication: linctl auth login --oauth",
				"Verify LINEAR_CLIENT_ID and LINEAR_CLIENT_SECRET are set",
			},
			Retryable: false,
		}
		return response
	}

	// Success - add metadata
	response.Success = true
	response.Data = map[string]interface{}{
		"authenticated": true,
		"method":        authStatus.Method,
		"user":          authStatus.User,
	}
	response.Metadata["auth_method"] = authStatus.Method
	response.Metadata["oauth_configured"] = authStatus.Method == "oauth"
	
	// Add actor configuration status
	actorConfig := oauth.LoadActorFromEnvironment()
	response.Metadata["actor_configured"] = actorConfig.IsConfigured()
	if actorConfig.IsConfigured() {
		response.Metadata["default_actor"] = actorConfig.DefaultActor
	}

	return response
}

// GetAgentStatus returns comprehensive status information for agents
func GetAgentStatus() *AgentResponse {
	response := &AgentResponse{
		Success:   true,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Metadata:  make(map[string]interface{}),
	}

	// Get authentication status
	authStatus, err := auth.GetAuthStatus()
	if err != nil {
		response.Success = false
		response.Error = &AgentError{
			Code:    "AUTH_STATUS_ERROR",
			Message: err.Error(),
			Retryable: true,
		}
		return response
	}

	// Get OAuth configuration
	oauthConfig := oauth.GetAgentConfiguration()
	
	// Get agent configuration
	agentConfig := LoadAgentConfig()

	response.Data = map[string]interface{}{
		"authentication": authStatus,
		"oauth":          oauthConfig,
		"agent_config":   agentConfig,
		"environment":    getEnvironmentSummary(),
	}

	// Add metadata for quick access
	response.Metadata["authenticated"] = authStatus.Authenticated
	response.Metadata["auth_method"] = authStatus.Method
	response.Metadata["oauth_configured"] = oauthConfig["oauth_configured"]
	response.Metadata["actor_configured"] = oauthConfig["actor_configured"]

	return response
}

// CreateStandardResponse creates a standardized response for agent operations
func CreateStandardResponse(success bool, data interface{}, err error) *AgentResponse {
	response := &AgentResponse{
		Success:   success,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Metadata:  make(map[string]interface{}),
	}

	if err != nil {
		response.Error = &AgentError{
			Code:      "OPERATION_ERROR",
			Message:   err.Error(),
			Retryable: isRetryableError(err),
		}
	}

	return response
}

// CreateErrorResponse creates a standardized error response for agents
func CreateErrorResponse(code, message string, retryable bool, suggestions ...string) *AgentResponse {
	return &AgentResponse{
		Success:   false,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Error: &AgentError{
			Code:        code,
			Message:     message,
			Suggestions: suggestions,
			Retryable:   retryable,
		},
		Metadata: make(map[string]interface{}),
	}
}

// FormatForAgent formats output appropriately for agent consumption
func FormatForAgent(data interface{}, jsonMode bool) (string, error) {
	if jsonMode {
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal JSON: %w", err)
		}
		return string(jsonData), nil
	}

	// For non-JSON mode, provide a simple string representation
	switch v := data.(type) {
	case *AgentResponse:
		if v.Success {
			return "SUCCESS", nil
		}
		return fmt.Sprintf("ERROR: %s", v.Error.Message), nil
	case string:
		return v, nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// ExitWithResponse exits with appropriate code and formatted output for agents
func ExitWithResponse(response *AgentResponse, jsonMode bool) {
	output, err := FormatForAgent(response, jsonMode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		os.Exit(1)
	}

	if jsonMode || response.Success {
		fmt.Println(output)
	} else {
		fmt.Fprintln(os.Stderr, output)
	}

	if response.Success {
		os.Exit(0)
	} else {
		// Use appropriate exit code based on error type
		if response.Error != nil {
			switch response.Error.Code {
			case "NOT_AUTHENTICATED", "OAUTH_CONFIG_ERROR":
				os.Exit(3) // Configuration error
			case "PERMISSION_DENIED":
				os.Exit(4) // Permission denied
			case "NOT_FOUND":
				os.Exit(5) // Resource not found
			default:
				os.Exit(1) // General error
			}
		}
		os.Exit(1)
	}
}

// Helper functions

func getBoolEnv(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	
	return boolValue
}

func getIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	
	return intValue
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := strings.ToLower(err.Error())
	
	// Network-related errors are typically retryable
	retryablePatterns := []string{
		"timeout",
		"connection",
		"network",
		"temporary",
		"rate limit",
		"503",
		"502",
		"500",
	}
	
	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}
	
	return false
}

func getEnvironmentSummary() map[string]interface{} {
	return map[string]interface{}{
		"LINEAR_CLIENT_ID":         os.Getenv("LINEAR_CLIENT_ID") != "",
		"LINEAR_CLIENT_SECRET":     os.Getenv("LINEAR_CLIENT_SECRET") != "",
		"LINEAR_DEFAULT_ACTOR":     os.Getenv("LINEAR_DEFAULT_ACTOR"),
		"LINEAR_DEFAULT_AVATAR_URL": os.Getenv("LINEAR_DEFAULT_AVATAR_URL"),
		"LINEAR_AGENT_SILENT":      getBoolEnv("LINEAR_AGENT_SILENT", false),
		"LINEAR_AGENT_JSON":        getBoolEnv("LINEAR_AGENT_JSON", false),
		"LINEAR_AGENT_TIMEOUT":     getIntEnv("LINEAR_AGENT_TIMEOUT", 30),
	}
}

// ActorOptions represents actor configuration for operations
type ActorOptions struct {
	Actor     string
	AvatarURL string
}

// ResolveActorOptions resolves actor options using provided values or environment defaults
func ResolveActorOptions(providedActor, providedAvatarURL string) *ActorOptions {
	actorConfig := oauth.LoadActorFromEnvironment()
	
	return &ActorOptions{
		Actor:     actorConfig.GetActor(providedActor),
		AvatarURL: actorConfig.GetAvatarURL(providedAvatarURL),
	}
}

// ValidateActorOptions validates actor options and provides suggestions
func ValidateActorOptions(options *ActorOptions) []string {
	var suggestions []string
	
	if options.Actor == "" {
		suggestions = append(suggestions, "Consider setting LINEAR_DEFAULT_ACTOR environment variable for consistent attribution")
	}
	
	if options.AvatarURL == "" {
		suggestions = append(suggestions, "Consider setting LINEAR_DEFAULT_AVATAR_URL environment variable for visual identification")
	}
	
	return suggestions
}