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

// GetAuthHeader returns the authorization header value
func GetAuthHeader() (string, error) {
	// First try OAuth with automatic token management
	if token, err := getValidOAuthToken(); err == nil && token != "" {
		return "Bearer " + token, nil
	}
	
	// Fall back to stored credentials
	config, err := loadAuth()
	if err != nil {
		return "", err
	}

	// Check OAuth token from config
	if config.OAuthToken != "" {
		return "Bearer " + config.OAuthToken, nil
	}

	// Fall back to API key for backward compatibility
	if config.APIKey != "" {
		return config.APIKey, nil
	}

	return "", fmt.Errorf("no valid authentication found")
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

// LoginWithOAuth handles OAuth authentication flow
func LoginWithOAuth(plaintext, jsonOut bool) error {
	// Try to load OAuth config from environment first
	oauthConfig, err := oauth.LoadFromEnvironment()
	if err != nil {
		return fmt.Errorf("failed to load OAuth config: %w", err)
	}
	
	// If environment variables are not set, prompt for them
	if !oauthConfig.IsComplete() {
		if !plaintext && !jsonOut {
			fmt.Println("\n" + color.New(color.FgYellow).Sprint("🔐 OAuth Authentication"))
			fmt.Println("You need Linear OAuth application credentials.")
			fmt.Println("Create an OAuth app at: https://linear.app/settings/api/applications/new")
			
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
	}

	if !oauthConfig.IsComplete() {
		return fmt.Errorf("OAuth client ID and secret are required")
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

	// Save the OAuth token to the legacy config for backward compatibility
	config := AuthConfig{
		OAuthToken: tokenResp.AccessToken,
	}
	err = saveAuth(config)
	if err != nil {
		return err
	}

	if !plaintext && !jsonOut {
		fmt.Printf("\n%s Authenticated via OAuth as %s (%s)\n",
			color.New(color.FgGreen).Sprint("✅"),
			color.New(color.FgCyan).Sprint(user.Name),
			color.New(color.FgCyan).Sprint(user.Email))
	}

	return nil
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

// RefreshOAuthToken forces a refresh of the OAuth token
func RefreshOAuthToken() error {
	// Try to load OAuth config from environment
	oauthConfig, err := oauth.LoadFromEnvironment()
	if err != nil {
		return fmt.Errorf("failed to load OAuth config: %w", err)
	}
	
	if !oauthConfig.IsComplete() {
		return fmt.Errorf("OAuth not configured via environment variables")
	}
	
	// Create OAuth client
	oauthClient, err := oauth.NewOAuthClientFromConfig(oauthConfig)
	if err != nil {
		return fmt.Errorf("failed to create OAuth client: %w", err)
	}
	
	// Force refresh token
	tokenResp, err := oauthClient.RefreshToken(context.Background(), oauthConfig.Scopes)
	if err != nil {
		return fmt.Errorf("failed to refresh OAuth token: %w", err)
	}
	
	// Update legacy config for backward compatibility
	config := AuthConfig{
		OAuthToken: tokenResp.AccessToken,
	}
	return saveAuth(config)
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
