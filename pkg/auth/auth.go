package auth

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dorkitude/linctl/pkg/api"
	"github.com/dorkitude/linctl/pkg/oauth"
	"github.com/fatih/color"
)

type User struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatarUrl,omitempty"`
}

type AuthConfig struct {
	APIKey     string `json:"api_key,omitempty"`
	OAuthToken string `json:"oauth_token,omitempty"`
}

// getConfigPath returns the path to the auth config file
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".linctl-auth.json"), nil
}

// saveAuth saves authentication credentials
func saveAuth(config AuthConfig) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

// loadAuth loads authentication credentials
func loadAuth() (*AuthConfig, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not authenticated")
		}
		return nil, err
	}

	var config AuthConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// GetAuthHeader returns the authorization header value with smart fallback and clear errors
func GetAuthHeader() (string, error) {
	// First try OAuth with automatic token management
	if token, err := getValidOAuthToken(); err == nil && token != "" {
		return "Bearer " + token, nil
	}
	
	// Fall back to stored credentials
	config, err := loadAuth()
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("not authenticated\n💡 Set up authentication: linctl auth login --oauth (recommended) or linctl auth login")
		}
		return "", fmt.Errorf("authentication error: %w\n💡 Try: linctl auth status", err)
	}

	// Check OAuth token from config
	if config.OAuthToken != "" {
		return "Bearer " + config.OAuthToken, nil
	}

	// Fall back to API key for backward compatibility
	if config.APIKey != "" {
		return config.APIKey, nil
	}

	return "", fmt.Errorf("no valid authentication found\n💡 Set up authentication: linctl auth login --oauth (recommended) or linctl auth login")
}

// getValidOAuthToken attempts to get a valid OAuth token using environment variables
func getValidOAuthToken() (string, error) {
	// Try to load OAuth config from environment
	oauthConfig, err := oauth.LoadFromEnvironment()
	if err != nil {
		return "", err
	}
	
	// Check if OAuth is configured
	if !oauthConfig.IsComplete() {
		return "", fmt.Errorf("OAuth not configured via environment variables")
	}
	
	// Create OAuth client
	oauthClient, err := oauth.NewOAuthClientFromConfig(oauthConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create OAuth client: %w", err)
	}
	
	// Get valid token (will use cached token if available and valid)
	tokenResp, err := oauthClient.GetValidToken(context.Background(), oauthConfig.Scopes)
	if err != nil {
		return "", fmt.Errorf("failed to get valid OAuth token: %w", err)
	}
	
	return tokenResp.AccessToken, nil
}

// Login handles the authentication flow
func Login(plaintext, jsonOut bool) error {
	return loginWithAPIKey(plaintext, jsonOut)
}

// loginWithAPIKey handles Personal API Key authentication
func loginWithAPIKey(plaintext, jsonOut bool) error {
	if !plaintext && !jsonOut {
		fmt.Println("\n" + color.New(color.FgYellow).Sprint("📝 Personal API Key Authentication"))
		fmt.Println("Get your API key from: https://linear.app/settings/api")

		// Get the config path to show to the user
		configPath, _ := getConfigPath()
		fmt.Printf("Your credentials will be stored in: %s\n", color.New(color.FgCyan).Sprint(configPath))
		fmt.Print("\nEnter your Personal API Key: ")
	}

	reader := bufio.NewReader(os.Stdin)
	apiKey, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Test the API key
	client := api.NewClient(apiKey)
	user, err := client.GetViewer(context.Background())
	if err != nil {
		return fmt.Errorf("invalid API key: %v", err)
	}

	// Save the API key
	config := AuthConfig{
		APIKey: apiKey,
	}
	err = saveAuth(config)
	if err != nil {
		return err
	}

	if !plaintext && !jsonOut {
		fmt.Printf("\n%s Authenticated as %s (%s)\n",
			color.New(color.FgGreen).Sprint("✅"),
			color.New(color.FgCyan).Sprint(user.Name),
			color.New(color.FgCyan).Sprint(user.Email))
	}

	return nil
}

// LoginWithOAuth handles OAuth authentication flow with existing auth detection
func LoginWithOAuth(plaintext, jsonOut bool) error {
	// Check for existing authentication
	existingConfig, _ := loadAuth()
	hasExistingAuth := existingConfig != nil && (existingConfig.APIKey != "" || existingConfig.OAuthToken != "")
	
	if hasExistingAuth && !plaintext && !jsonOut {
		if existingConfig.APIKey != "" {
			fmt.Println(color.New(color.FgBlue).Sprint("ℹ️  Detected existing API key authentication"))
			fmt.Println(color.New(color.FgBlue).Sprint("🔄 Setting up OAuth (API key will remain as fallback)"))
		} else {
			fmt.Println(color.New(color.FgBlue).Sprint("ℹ️  Updating existing OAuth authentication"))
		}
	}

	// Try to load OAuth config from environment first
	oauthConfig, err := oauth.LoadFromEnvironment()
	if err != nil {
		return fmt.Errorf("failed to load OAuth config: %w", err)
	}
	
	// If environment variables are not set, prompt for them
	if !oauthConfig.IsComplete() {
		if !plaintext && !jsonOut {
			fmt.Println("\n" + color.New(color.FgYellow).Sprint("🔐 OAuth Authentication Setup"))
			fmt.Println("You need Linear OAuth application credentials.")
			fmt.Println("Create an OAuth app at: https://linear.app/settings/api/applications/new")
			fmt.Println()
			fmt.Println(color.New(color.FgCyan).Sprint("💡 Tip: Set LINEAR_CLIENT_ID and LINEAR_CLIENT_SECRET environment variables for automated workflows"))
			
			// Get the config path to show to the user
			configPath, _ := getConfigPath()
			fmt.Printf("Your credentials will be stored in: %s\n", color.New(color.FgCyan).Sprint(configPath))
		}

		if oauthConfig.ClientID == "" {
			if !plaintext && !jsonOut {
				fmt.Print("\nEnter your OAuth Client ID: ")
			}
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			oauthConfig.ClientID = strings.TrimSpace(input)
		}

		if oauthConfig.ClientSecret == "" {
			if !plaintext && !jsonOut {
				fmt.Print("Enter your OAuth Client Secret: ")
			}
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			oauthConfig.ClientSecret = strings.TrimSpace(input)
		}
	} else if !plaintext && !jsonOut {
		fmt.Println(color.New(color.FgGreen).Sprint("✅ Using OAuth configuration from environment variables"))
	}

	if !oauthConfig.IsComplete() {
		return fmt.Errorf("OAuth client ID and secret are required")
	}

	if !plaintext && !jsonOut {
		fmt.Println(color.New(color.FgYellow).Sprint("🌐 Authenticating with Linear OAuth..."))
	}

	// Create OAuth client and get access token
	oauthClient, err := oauth.NewOAuthClientFromConfig(oauthConfig)
	if err != nil {
		return fmt.Errorf("failed to create OAuth client: %w", err)
	}
	
	tokenResp, err := oauthClient.GetValidToken(context.Background(), oauthConfig.Scopes)
	if err != nil {
		return fmt.Errorf("failed to get OAuth token: %v", err)
	}

	// Test the token by getting current user
	client := api.NewClient("Bearer " + tokenResp.AccessToken)
	user, err := client.GetViewer(context.Background())
	if err != nil {
		return fmt.Errorf("failed to validate OAuth token: %v", err)
	}

	// Preserve existing API key if present, add OAuth token
	config := AuthConfig{
		OAuthToken: tokenResp.AccessToken,
	}
	if existingConfig != nil && existingConfig.APIKey != "" {
		config.APIKey = existingConfig.APIKey
	}
	
	err = saveAuth(config)
	if err != nil {
		return err
	}

	if !plaintext && !jsonOut {
		fmt.Printf("\n%s OAuth setup complete! Future commands will use OAuth automatically.\n",
			color.New(color.FgGreen).Sprint("✅"))
		fmt.Printf("Authenticated as: %s (%s)\n",
			color.New(color.FgCyan).Sprint(user.Name),
			color.New(color.FgCyan).Sprint(user.Email))
		
		if existingConfig != nil && existingConfig.APIKey != "" {
			fmt.Println(color.New(color.FgBlue).Sprint("💡 Your API key is preserved as a fallback"))
		}
	}

	return nil
}

// AuthStatus represents comprehensive authentication status
type AuthStatus struct {
	Authenticated bool              `json:"authenticated"`
	Method        string            `json:"method"`        // "oauth", "api_key", or "none"
	User          *User             `json:"user,omitempty"`
	TokenExpiry   *string           `json:"token_expires_at,omitempty"`
	Scopes        []string          `json:"scopes,omitempty"`
	Suggestions   []string          `json:"suggestions,omitempty"`
	Environment   map[string]interface{} `json:"environment,omitempty"`
}

// GetAuthMethod returns the current authentication method
func GetAuthMethod() (string, error) {
	config, err := loadAuth()
	if err != nil {
		return "", err
	}

	if config.OAuthToken != "" {
		return "oauth", nil
	}

	if config.APIKey != "" {
		return "api_key", nil
	}

	return "", fmt.Errorf("no valid authentication found")
}

// GetAuthStatus returns comprehensive authentication status with guidance
func GetAuthStatus() (*AuthStatus, error) {
	status := &AuthStatus{
		Authenticated: false,
		Method:        "none",
		Suggestions:   []string{},
	}

	// Try to get current user to determine authentication status
	user, userErr := GetCurrentUser()
	if userErr == nil {
		status.Authenticated = true
		status.User = user
	}

	// Determine authentication method and add relevant information
	authMethod, methodErr := GetAuthMethod()
	if methodErr == nil {
		status.Method = authMethod
	}

	// Get OAuth information if available
	oauthInfo, oauthErr := GetOAuthTokenInfo()
	if oauthErr == nil && oauthInfo["configured"].(bool) {
		status.Environment = oauthInfo["environment"].(map[string]interface{})
		
		if status.Method == "oauth" {
			// Add OAuth-specific information
			if valid, ok := oauthInfo["valid"].(bool); ok && valid {
				if expiresAt, ok := oauthInfo["expires_at"].(string); ok {
					status.TokenExpiry = &expiresAt
				}
				if scope, ok := oauthInfo["scope"].(string); ok && scope != "" {
					status.Scopes = strings.Split(scope, " ")
				}
			} else {
				status.Suggestions = append(status.Suggestions, "OAuth token is expired or invalid. Refresh with: linctl auth refresh")
			}
		}
	}

	// Add intelligent suggestions based on current state
	if !status.Authenticated {
		status.Suggestions = append(status.Suggestions, "Set up authentication with: linctl auth login --oauth (recommended) or linctl auth login")
	} else if status.Method == "api_key" {
		// Check if OAuth is configured via environment
		if oauthErr == nil && oauthInfo["configured"].(bool) {
			status.Suggestions = append(status.Suggestions, "OAuth is configured via environment variables. Switch to OAuth for enhanced features: linctl auth login --oauth")
		} else {
			status.Suggestions = append(status.Suggestions, "Consider upgrading to OAuth for enhanced features like actor attribution: linctl auth login --oauth")
		}
	}

	// Add environment configuration guidance
	if oauthErr == nil && !oauthInfo["configured"].(bool) {
		status.Suggestions = append(status.Suggestions, "For automated workflows, configure OAuth via environment variables: LINEAR_CLIENT_ID and LINEAR_CLIENT_SECRET")
	}

	return status, nil
}

// GetCurrentUser returns the current authenticated user
func GetCurrentUser() (*User, error) {
	authHeader, err := GetAuthHeader()
	if err != nil {
		return nil, err
	}

	client := api.NewClient(authHeader)
	apiUser, err := client.GetViewer(context.Background())
	if err != nil {
		return nil, err
	}

	// Convert api.User to auth.User
	return &User{
		ID:        apiUser.ID,
		Name:      apiUser.Name,
		Email:     apiUser.Email,
		AvatarURL: apiUser.AvatarURL,
	}, nil
}

// RefreshOAuthTokenWithFeedback forces a refresh of the OAuth token with user-friendly errors
func RefreshOAuthTokenWithFeedback() error {
	// Try to load OAuth config from environment
	oauthConfig, err := oauth.LoadFromEnvironment()
	if err != nil {
		return fmt.Errorf("OAuth configuration error: %w\n💡 Ensure LINEAR_CLIENT_ID and LINEAR_CLIENT_SECRET are set", err)
	}
	
	if !oauthConfig.IsComplete() {
		return fmt.Errorf("OAuth not configured via environment variables\n💡 Set LINEAR_CLIENT_ID and LINEAR_CLIENT_SECRET, then try: linctl auth login --oauth")
	}
	
	// Create OAuth client
	oauthClient, err := oauth.NewOAuthClientFromConfig(oauthConfig)
	if err != nil {
		return fmt.Errorf("failed to create OAuth client: %w\n💡 Check your OAuth configuration and try again", err)
	}
	
	// Force refresh token
	tokenResp, err := oauthClient.RefreshToken(context.Background(), oauthConfig.Scopes)
	if err != nil {
		return fmt.Errorf("OAuth token expired and refresh failed\n💡 Please re-authenticate: linctl auth login --oauth")
	}
	
	// Preserve existing API key if present
	existingConfig, _ := loadAuth()
	config := AuthConfig{
		OAuthToken: tokenResp.AccessToken,
	}
	if existingConfig != nil && existingConfig.APIKey != "" {
		config.APIKey = existingConfig.APIKey
	}
	
	return saveAuth(config)
}

// RefreshOAuthToken forces a refresh of the OAuth token (backward compatibility)
func RefreshOAuthToken() error {
	return RefreshOAuthTokenWithFeedback()
}

// GetOAuthTokenInfo returns information about the current OAuth token
func GetOAuthTokenInfo() (map[string]interface{}, error) {
	// Try to load OAuth config from environment
	oauthConfig, err := oauth.LoadFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to load OAuth config: %w", err)
	}
	
	if !oauthConfig.IsComplete() {
		return map[string]interface{}{
			"configured": false,
			"error":      "OAuth not configured via environment variables",
		}, nil
	}
	
	// Create OAuth client
	oauthClient, err := oauth.NewOAuthClientFromConfig(oauthConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth client: %w", err)
	}
	
	// Get token info
	tokenInfo := oauthClient.GetStoredTokenInfo()
	tokenInfo["configured"] = true
	tokenInfo["environment"] = oauth.GetEnvironmentStatus()
	
	return tokenInfo, nil
}

// Logout clears stored credentials
func Logout() error {
	// Clear legacy config
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	err = os.Remove(configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Clear OAuth token store
	tokenStore, err := oauth.NewTokenStore()
	if err == nil {
		// Ignore error if token store doesn't exist
		_ = tokenStore.ClearToken()
	}

	return nil
}
