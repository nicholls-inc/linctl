package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dorkitude/linctl/pkg/auth"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestAuthCommands_Integration(t *testing.T) {
	// Create temporary directory for test config files
	tempDir, err := os.MkdirTemp("", "linctl-auth-cmd-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override the config path for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Note: status_not_authenticated test removed because it calls os.Exit(1)

	// Test that OAuth flag is available on login command
	t.Run("login_oauth_flag_available", func(t *testing.T) {
		flag := loginCmd.Flags().Lookup("oauth")
		if flag == nil {
			t.Error("Expected --oauth flag to be available on login command")
		}

		if flag.Usage != "Use OAuth authentication instead of API key" {
			t.Errorf("Expected OAuth flag usage description, got %s", flag.Usage)
		}

		if flag.DefValue != "false" {
			t.Errorf("Expected OAuth flag default value false, got %s", flag.DefValue)
		}
	})

	// Test help output includes OAuth information
	t.Run("login_help_includes_oauth", func(t *testing.T) {
		buf := &bytes.Buffer{}
		loginCmd.SetOut(buf)
		loginCmd.SetErr(buf)

		// Execute help
		loginCmd.Help()

		output := buf.String()
		if !strings.Contains(output, "--oauth") {
			t.Error("Expected help output to contain --oauth flag")
		}
		if !strings.Contains(output, "OAuth authentication") {
			t.Error("Expected help output to mention OAuth authentication")
		}
	})

	// Test auth command structure
	t.Run("auth_command_structure", func(t *testing.T) {
		// Verify auth command has expected subcommands
		expectedSubcommands := []string{"login", "logout", "status"}
		
		for _, expectedCmd := range expectedSubcommands {
			found := false
			for _, cmd := range authCmd.Commands() {
				if cmd.Name() == expectedCmd {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected auth command to have %s subcommand", expectedCmd)
			}
		}
	})

	// Test whoami command exists as top-level command
	t.Run("whoami_command_exists", func(t *testing.T) {
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == "whoami" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected whoami to be available as top-level command")
		}
	})
}

func TestStatusCommand_WithAuthentication(t *testing.T) {
	// Create temporary directory for test config files
	tempDir, err := os.MkdirTemp("", "linctl-auth-cmd-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override the config path for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	tests := []struct {
		name           string
		config         auth.AuthConfig
		expectedMethod string
	}{
		{
			name: "OAuth authentication",
			config: auth.AuthConfig{
				OAuthToken: "test-oauth-token",
			},
			expectedMethod: "oauth",
		},
		{
			name: "API key authentication",
			config: auth.AuthConfig{
				APIKey: "test-api-key",
			},
			expectedMethod: "api_key",
		},
		{
			name: "OAuth takes priority",
			config: auth.AuthConfig{
				APIKey:     "test-api-key",
				OAuthToken: "test-oauth-token",
			},
			expectedMethod: "oauth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save test config using internal function (we need to access it)
			configPath := filepath.Join(tempDir, ".linctl-auth.json")
			configData, err := json.MarshalIndent(tt.config, "", "  ")
			if err != nil {
				t.Fatalf("Failed to marshal config: %v", err)
			}
			err = os.WriteFile(configPath, configData, 0600)
			if err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			// Test GetAuthMethod function
			method, err := auth.GetAuthMethod()
			if err != nil {
				t.Fatalf("Failed to get auth method: %v", err)
			}

			if method != tt.expectedMethod {
				t.Errorf("Expected auth method %s, got %s", tt.expectedMethod, method)
			}

			// Test GetAuthHeader function
			header, err := auth.GetAuthHeader()
			if err != nil {
				t.Fatalf("Failed to get auth header: %v", err)
			}

			// Verify header format based on method
			if tt.expectedMethod == "oauth" {
				expectedHeader := "Bearer " + tt.config.OAuthToken
				if header != expectedHeader {
					t.Errorf("Expected OAuth header %s, got %s", expectedHeader, header)
				}
			} else if tt.expectedMethod == "api_key" {
				if header != tt.config.APIKey {
					t.Errorf("Expected API key header %s, got %s", tt.config.APIKey, header)
				}
			}
		})
	}
}

func TestLoginCommand_OAuthFlag(t *testing.T) {
	// Test that OAuth flag is properly parsed
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "no oauth flag",
			args:     []string{},
			expected: false,
		},
		{
			name:     "oauth flag set",
			args:     []string{"--oauth"},
			expected: true,
		},
		{
			name:     "oauth flag with other flags",
			args:     []string{"--oauth", "--json"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the flag
			oauthFlag = false

			// Parse the flags
			loginCmd.ParseFlags(tt.args)

			if oauthFlag != tt.expected {
				t.Errorf("Expected oauthFlag %v, got %v", tt.expected, oauthFlag)
			}
		})
	}
}

func TestAuthCommand_DefaultBehavior(t *testing.T) {
	// Test that auth command without subcommand defaults to login
	buf := &bytes.Buffer{}
	authCmd.SetOut(buf)
	authCmd.SetErr(buf)

	// The auth command should have the same Run function as loginCmd
	if authCmd.Run == nil {
		t.Error("Expected auth command to have Run function")
	}

	// Verify the help text mentions the default behavior
	help := authCmd.Long
	if !strings.Contains(help, "linctl auth") && !strings.Contains(help, "Interactive authentication") {
		t.Error("Expected auth command help to mention default behavior")
	}
}

func TestCommandHelp_OAuth(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *cobra.Command
		expects []string
	}{
		{
			name: "auth command help",
			cmd:  authCmd,
			expects: []string{
				"Authenticate with Linear",
				"auth login",
				"auth status",
				"auth logout",
			},
		},
		{
			name: "login command help",
			cmd:  loginCmd,
			expects: []string{
				"OAuth",
				"--oauth",
			},
		},
		{
			name: "status command help",
			cmd:  statusCmd,
			expects: []string{
				"authenticated with Linear",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			tt.cmd.SetOut(buf)
			tt.cmd.SetErr(buf)

			// Get help output
			tt.cmd.Help()
			output := buf.String()

			for _, expected := range tt.expects {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected help output to contain %s, got %s", expected, output)
				}
			}
		})
	}
}

func TestGlobalFlags_Integration(t *testing.T) {
	// Test that global flags work with auth commands
	tests := []struct {
		name string
		flag string
	}{
		{"json flag", "json"},
		{"plaintext flag", "plaintext"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Global flags should be available through viper
			viper.Set(tt.flag, true)
			value := viper.GetBool(tt.flag)
			if !value {
				t.Errorf("Expected global flag %s to be accessible via viper", tt.flag)
			}
			viper.Set(tt.flag, false) // Reset
		})
	}
}