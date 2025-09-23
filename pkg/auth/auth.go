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
	config, err := loadAuth()
	if err != nil {
		return "", err
	}

	// Check OAuth token first
	if config.OAuthToken != "" {
		return "Bearer " + config.OAuthToken, nil
	}

	// Fall back to API key for backward compatibility
	if config.APIKey != "" {
		return config.APIKey, nil
	}

	return "", fmt.Errorf("no valid authentication found")
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
	// Check for environment variables first
	clientID := os.Getenv("LINEAR_CLIENT_ID")
	clientSecret := os.Getenv("LINEAR_CLIENT_SECRET")
	baseURL := os.Getenv("LINEAR_BASE_URL")
	
	if baseURL == "" {
		baseURL = "https://api.linear.app"
	}

	// If environment variables are not set, prompt for them
	if clientID == "" || clientSecret == "" {
		if !plaintext && !jsonOut {
			fmt.Println("\n" + color.New(color.FgYellow).Sprint("🔐 OAuth Authentication"))
			fmt.Println("You need Linear OAuth application credentials.")
			fmt.Println("Create an OAuth app at: https://linear.app/settings/api/applications/new")
			
			// Get the config path to show to the user
			configPath, _ := getConfigPath()
			fmt.Printf("Your credentials will be stored in: %s\n", color.New(color.FgCyan).Sprint(configPath))
		}

		if clientID == "" {
			if !plaintext && !jsonOut {
				fmt.Print("\nEnter your OAuth Client ID: ")
			}
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			clientID = strings.TrimSpace(input)
		}

		if clientSecret == "" {
			if !plaintext && !jsonOut {
				fmt.Print("Enter your OAuth Client Secret: ")
			}
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			clientSecret = strings.TrimSpace(input)
		}
	}

	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("OAuth client ID and secret are required")
	}

	// Create OAuth client and get access token
	oauthClient := oauth.NewOAuthClient(clientID, clientSecret, baseURL)
	scopes := []string{"read", "write", "issues:create", "comments:create"}
	
	tokenResp, err := oauthClient.GetAccessToken(context.Background(), scopes)
	if err != nil {
		return fmt.Errorf("failed to get OAuth token: %v", err)
	}

	// Test the token by getting current user
	client := api.NewClient("Bearer " + tokenResp.AccessToken)
	user, err := client.GetViewer(context.Background())
	if err != nil {
		return fmt.Errorf("failed to validate OAuth token: %v", err)
	}

	// Save the OAuth token
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

// Logout clears stored credentials
func Logout() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	err = os.Remove(configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
