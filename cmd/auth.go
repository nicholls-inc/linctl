package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/nicholls-inc/linctl/pkg/auth"
	"github.com/nicholls-inc/linctl/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var oauthFlag bool

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Linear",
	Long: `Authenticate with Linear using Personal API Key.

Examples:
  linctl auth              # Interactive authentication
  linctl auth login        # Same as above
  linctl auth status       # Check authentication status
  linctl auth logout       # Clear stored credentials`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior is to run login
		loginCmd.Run(cmd, args)
	},
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Linear",
	Long:  `Authenticate with Linear using Personal API Key or OAuth.`,
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		if !plaintext && !jsonOut {
			fmt.Println(color.New(color.FgCyan, color.Bold).Sprint("🔐 Linear Authentication"))
			fmt.Println()
		}

		var err error
		if oauthFlag {
			err = auth.LoginWithOAuth(plaintext, jsonOut)
		} else {
			err = auth.Login(plaintext, jsonOut)
		}

		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if !plaintext && !jsonOut {
			fmt.Println(color.New(color.FgGreen).Sprint("✅ Successfully authenticated with Linear!"))
		} else if jsonOut {
			output.JSON(map[string]interface{}{
				"status":  "success",
				"message": "Successfully authenticated with Linear",
			})
		} else {
			fmt.Println("Successfully authenticated with Linear")
		}
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Long:  `Check if you are currently authenticated with Linear and get helpful guidance.`,
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		status, err := auth.GetAuthStatus()
		if err != nil {
			if jsonOut {
				output.JSON(map[string]interface{}{
					"authenticated": false,
					"error":         err.Error(),
				})
			} else {
				output.Error(fmt.Sprintf("Failed to get auth status: %v", err), plaintext, jsonOut)
			}
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(status)
			return
		}

		if !status.Authenticated {
			if plaintext {
				fmt.Println("Not authenticated")
				for _, suggestion := range status.Suggestions {
					fmt.Printf("Suggestion: %s\n", suggestion)
				}
			} else {
				fmt.Println(color.New(color.FgRed).Sprint("❌ Not authenticated"))
				fmt.Println()
				for _, suggestion := range status.Suggestions {
					fmt.Printf("%s %s\n", color.New(color.FgBlue).Sprint("💡"), suggestion)
				}
			}
			os.Exit(1)
		}

		// Authenticated - show status
		if plaintext {
			fmt.Printf("Authenticated as: %s (%s) via %s\n", status.User.Name, status.User.Email, status.Method)
			if status.TokenExpiry != nil {
				fmt.Printf("Token expires: %s\n", *status.TokenExpiry)
			}
			if len(status.Scopes) > 0 {
				fmt.Printf("Scopes: %s\n", strings.Join(status.Scopes, ", "))
			}
			for _, suggestion := range status.Suggestions {
				fmt.Printf("Suggestion: %s\n", suggestion)
			}
		} else {
			// Enhanced colorful output
			fmt.Println(color.New(color.FgGreen).Sprint("✅ Authenticated"))

			// Show method with appropriate icon
			methodIcon := "🔑"
			if status.Method == "oauth" {
				methodIcon = "🔐"
			}
			fmt.Printf("%s Method: %s\n", methodIcon, color.New(color.FgCyan).Sprint(status.Method))

			// User info
			fmt.Printf("👤 User: %s (%s)\n",
				color.New(color.FgCyan).Sprint(status.User.Name),
				color.New(color.FgCyan).Sprint(status.User.Email))

			// Token expiry for OAuth
			if status.TokenExpiry != nil {
				fmt.Printf("🔑 Token expires: %s\n", color.New(color.FgYellow).Sprint(*status.TokenExpiry))
			}

			// Scopes for OAuth
			if len(status.Scopes) > 0 {
				fmt.Printf("📋 Scopes: %s\n", color.New(color.FgCyan).Sprint(strings.Join(status.Scopes, ", ")))
			}

			// Show suggestions if any
			if len(status.Suggestions) > 0 {
				fmt.Println()
				for _, suggestion := range status.Suggestions {
					fmt.Printf("%s %s\n", color.New(color.FgBlue).Sprint("💡"), suggestion)
				}
			}
		}
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from Linear",
	Long:  `Clear stored Linear credentials.`,
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		err := auth.Logout()
		if err != nil {
			output.Error(fmt.Sprintf("Logout failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(map[string]interface{}{
				"status":  "success",
				"message": "Successfully logged out",
			})
		} else if plaintext {
			fmt.Println("Successfully logged out")
		} else {
			fmt.Println(color.New(color.FgGreen).Sprint("✅ Successfully logged out"))
		}
	},
}

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh OAuth token",
	Long:  `Force refresh of the OAuth access token.`,
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		if !plaintext && !jsonOut {
			fmt.Println(color.New(color.FgYellow).Sprint("🔄 Refreshing OAuth token..."))
		}

		err := auth.RefreshOAuthTokenWithFeedback()
		if err != nil {
			if jsonOut {
				output.JSON(map[string]interface{}{
					"status": "error",
					"error":  err.Error(),
				})
			} else if plaintext {
				fmt.Printf("Token refresh failed: %v\n", err)
			} else {
				fmt.Printf("%s %v\n", color.New(color.FgRed).Sprint("❌"), err)
			}
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(map[string]interface{}{
				"status":  "success",
				"message": "OAuth token refreshed successfully",
			})
		} else if plaintext {
			fmt.Println("OAuth token refreshed successfully")
		} else {
			fmt.Println(color.New(color.FgGreen).Sprint("✅ OAuth token refreshed successfully"))
		}
	},
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current user",
	Long:  `Display information about the currently authenticated user.`,
	Run: func(cmd *cobra.Command, args []string) {
		statusCmd.Run(cmd, args)
	},
}

var authAgentStatusCmd = &cobra.Command{
	Use:   "agent-status",
	Short: "Show agent-optimized status",
	Long:  `Display comprehensive status information optimized for agent workflows, including OAuth configuration, environment variables, and actor settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		// Import agent package for this command
		// Note: This would require importing the agent package
		// For now, we'll provide a simplified version using existing auth functions

		status, err := auth.GetAuthStatus()
		if err != nil {
			if jsonOut {
				output.JSON(map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				})
			} else {
				output.Error(fmt.Sprintf("Failed to get status: %v", err), plaintext, jsonOut)
			}
			os.Exit(1)
		}

		// Get OAuth token info for additional details
		oauthInfo, _ := auth.GetOAuthTokenInfo()

		// Enhanced status for agents
		agentStatus := map[string]interface{}{
			"success":          status.Authenticated,
			"authenticated":    status.Authenticated,
			"method":           status.Method,
			"user":             status.User,
			"token_expires_at": status.TokenExpiry,
			"scopes":           status.Scopes,
			"suggestions":      status.Suggestions,
			"environment":      status.Environment,
			"oauth_info":       oauthInfo,
			"timestamp":        time.Now().UTC().Format(time.RFC3339),
		}

		if jsonOut {
			output.JSON(agentStatus)
		} else {
			// Provide human-readable output for non-JSON mode
			if status.Authenticated {
				if plaintext {
					fmt.Printf("Authenticated: true\n")
					fmt.Printf("Method: %s\n", status.Method)
					if status.User != nil {
						fmt.Printf("User: %s (%s)\n", status.User.Name, status.User.Email)
					}
				} else {
					fmt.Println(color.New(color.FgGreen).Sprint("✅ Agent Status: Ready"))
					fmt.Printf("🔐 Method: %s\n", color.New(color.FgCyan).Sprint(status.Method))
					if status.User != nil {
						fmt.Printf("👤 User: %s (%s)\n",
							color.New(color.FgCyan).Sprint(status.User.Name),
							color.New(color.FgCyan).Sprint(status.User.Email))
					}
				}
			} else {
				if plaintext {
					fmt.Printf("Authenticated: false\n")
				} else {
					fmt.Println(color.New(color.FgRed).Sprint("❌ Agent Status: Not Ready"))
				}
			}

			// Show suggestions
			for _, suggestion := range status.Suggestions {
				if plaintext {
					fmt.Printf("Suggestion: %s\n", suggestion)
				} else {
					fmt.Printf("%s %s\n", color.New(color.FgBlue).Sprint("💡"), suggestion)
				}
			}
		}

		if !status.Authenticated {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(statusCmd)
	authCmd.AddCommand(refreshCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(authAgentStatusCmd)

	// Add OAuth flag to login command
	loginCmd.Flags().BoolVar(&oauthFlag, "oauth", false, "Use OAuth authentication instead of API key")

	// Add whoami as a top-level command too
	rootCmd.AddCommand(whoamiCmd)
}
