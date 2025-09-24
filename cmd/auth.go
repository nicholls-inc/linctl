package cmd

import (
	"fmt"
	"os"

	"github.com/dorkitude/linctl/pkg/auth"
	"github.com/dorkitude/linctl/pkg/output"
	"github.com/fatih/color"
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
	Long:  `Check if you are currently authenticated with Linear.`,
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		user, err := auth.GetCurrentUser()
		if err != nil {
			if !plaintext && !jsonOut {
				fmt.Println(color.New(color.FgRed).Sprint("❌ Not authenticated"))
			} else if jsonOut {
				output.JSON(map[string]interface{}{
					"authenticated": false,
					"error":         err.Error(),
				})
			} else {
				fmt.Println("Not authenticated")
			}
			os.Exit(1)
		}

		authMethod, _ := auth.GetAuthMethod()

		// Get OAuth token info if using OAuth
		var oauthInfo map[string]interface{}
		if authMethod == "oauth" {
			oauthInfo, _ = auth.GetOAuthTokenInfo()
		}

		if jsonOut {
			result := map[string]interface{}{
				"authenticated": true,
				"method":        authMethod,
				"user":          user,
			}
			if oauthInfo != nil {
				result["oauth"] = oauthInfo
			}
			output.JSON(result)
		} else if plaintext {
			fmt.Printf("Authenticated as: %s (%s) via %s\n", user.Name, user.Email, authMethod)
			if oauthInfo != nil && oauthInfo["valid"].(bool) {
				if expiresIn, ok := oauthInfo["expires_in_human"].(string); ok {
					fmt.Printf("Token expires in: %s\n", expiresIn)
				}
			}
		} else {
			fmt.Println(color.New(color.FgGreen).Sprint("✅ Authenticated"))
			fmt.Printf("Method: %s\n", color.New(color.FgCyan).Sprint(authMethod))
			fmt.Printf("User: %s\n", color.New(color.FgCyan).Sprint(user.Name))
			fmt.Printf("Email: %s\n", color.New(color.FgCyan).Sprint(user.Email))
			
			if oauthInfo != nil {
				if valid, ok := oauthInfo["valid"].(bool); ok && valid {
					if expiresIn, ok := oauthInfo["expires_in_human"].(string); ok {
						fmt.Printf("Token expires in: %s\n", color.New(color.FgYellow).Sprint(expiresIn))
					}
					if scope, ok := oauthInfo["scope"].(string); ok {
						fmt.Printf("Scopes: %s\n", color.New(color.FgCyan).Sprint(scope))
					}
				} else {
					fmt.Println(color.New(color.FgRed).Sprint("⚠️  OAuth token is expired or invalid"))
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

		err := auth.RefreshOAuthToken()
		if err != nil {
			output.Error(fmt.Sprintf("Token refresh failed: %v", err), plaintext, jsonOut)
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

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(statusCmd)
	authCmd.AddCommand(refreshCmd)
	authCmd.AddCommand(logoutCmd)

	// Add OAuth flag to login command
	loginCmd.Flags().BoolVar(&oauthFlag, "oauth", false, "Use OAuth authentication instead of API key")

	// Add whoami as a top-level command too
	rootCmd.AddCommand(whoamiCmd)
}
