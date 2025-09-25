package cmd

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/nicholls-inc/linctl/pkg/agent"
	"github.com/nicholls-inc/linctl/pkg/auth"
	"github.com/nicholls-inc/linctl/pkg/oauth"
	"github.com/nicholls-inc/linctl/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// agentCmd represents the agent command
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Agent-optimized commands and utilities",
	Long: `Commands and utilities specifically designed for AI agents and automated workflows.

These commands provide:
- JSON output for all operations
- Proper exit codes for script automation
- Environment variable validation
- Comprehensive status information
- Silent operation modes

Examples:
  linctl agent validate     # Validate environment for agent workflows
  linctl agent status       # Get comprehensive agent status
  linctl agent config       # Show agent configuration`,
}

var agentValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate environment for agent workflows",
	Long: `Validate that the environment is properly configured for agent workflows.

This command checks:
- OAuth configuration (LINEAR_CLIENT_ID, LINEAR_CLIENT_SECRET)
- Authentication status
- Actor configuration (LINEAR_DEFAULT_ACTOR, LINEAR_DEFAULT_AVATAR_URL)
- Network connectivity

Exit codes:
  0 - Environment is valid and ready for agent workflows
  1 - General validation error
  3 - Configuration error (missing environment variables)`,
	Run: func(cmd *cobra.Command, args []string) {
		jsonOut := viper.GetBool("json")

		// Always use JSON mode for agent commands unless explicitly disabled
		if !cmd.Flags().Changed("json") && !viper.IsSet("json") {
			jsonOut = true
		}

		response := agent.ValidateAgentEnvironment()
		agent.ExitWithResponse(response, jsonOut)
	},
}

var agentStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get comprehensive agent status",
	Long: `Get comprehensive status information optimized for agent consumption.

Returns detailed information about:
- Authentication status and method
- OAuth configuration
- Environment variables
- Actor configuration
- Token expiry and scopes

Output is always in JSON format for easy parsing by agents.`,
	Run: func(cmd *cobra.Command, args []string) {
		jsonOut := viper.GetBool("json")

		// Always use JSON mode for agent commands unless explicitly disabled
		if !cmd.Flags().Changed("json") && !viper.IsSet("json") {
			jsonOut = true
		}

		response := agent.GetAgentStatus()
		agent.ExitWithResponse(response, jsonOut)
	},
}

var agentConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Show agent configuration",
	Long: `Display current agent configuration including environment variables and defaults.

Shows:
- OAuth configuration status
- Actor configuration
- Environment variable status
- Agent-specific settings`,
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		// Always use JSON mode for agent commands unless explicitly disabled
		if !cmd.Flags().Changed("json") && !viper.IsSet("json") {
			jsonOut = true
		}

		// Get configuration information
		oauthConfig := oauth.GetAgentConfiguration()
		agentConfig := agent.LoadAgentConfig()
		envStatus := oauth.GetEnvironmentStatus()

		configData := map[string]interface{}{
			"success":     true,
			"oauth":       oauthConfig,
			"agent":       agentConfig,
			"environment": envStatus,
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
		}

		if jsonOut {
			output.JSON(configData)
		} else {
			// Human-readable output
			if plaintext {
				fmt.Printf("OAuth Configured: %v\n", oauthConfig["oauth_configured"])
				fmt.Printf("Actor Configured: %v\n", oauthConfig["actor_configured"])
				fmt.Printf("Default Actor: %s\n", agentConfig.DefaultActor)
				fmt.Printf("Default Avatar URL: %s\n", agentConfig.DefaultAvatarURL)
			} else {
				fmt.Println(color.New(color.FgCyan, color.Bold).Sprint("🤖 Agent Configuration"))
				fmt.Println()

				// OAuth status
				if oauthConfig["oauth_configured"].(bool) {
					fmt.Printf("%s OAuth: %s\n",
						color.New(color.FgGreen).Sprint("✅"),
						color.New(color.FgGreen).Sprint("Configured"))
				} else {
					fmt.Printf("%s OAuth: %s\n",
						color.New(color.FgRed).Sprint("❌"),
						color.New(color.FgRed).Sprint("Not Configured"))
				}

				// Actor status
				if oauthConfig["actor_configured"].(bool) {
					fmt.Printf("%s Actor: %s\n",
						color.New(color.FgGreen).Sprint("✅"),
						color.New(color.FgGreen).Sprint("Configured"))
					if agentConfig.DefaultActor != "" {
						fmt.Printf("  Default Actor: %s\n", color.New(color.FgCyan).Sprint(agentConfig.DefaultActor))
					}
				} else {
					fmt.Printf("%s Actor: %s\n",
						color.New(color.FgYellow).Sprint("⚠️"),
						color.New(color.FgYellow).Sprint("Not Configured"))
					fmt.Printf("  %s Set LINEAR_DEFAULT_ACTOR for consistent attribution\n",
						color.New(color.FgBlue).Sprint("💡"))
				}

				// Environment variables
				fmt.Println()
				fmt.Println(color.New(color.FgCyan).Sprint("Environment Variables:"))
				for key, value := range envStatus {
					fmt.Printf("  %s: %v\n", key, value)
				}
			}
		}
	},
}

var agentTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test agent functionality",
	Long: `Test basic agent functionality including authentication and API operations.

This command performs a series of tests:
1. Environment validation
2. Authentication check
3. Basic API operation (get viewer)
4. Actor attribution test (if configured)

Useful for verifying agent setup before running automated workflows.`,
	Run: func(cmd *cobra.Command, args []string) {
		jsonOut := viper.GetBool("json")

		// Always use JSON mode for agent commands unless explicitly disabled
		if !cmd.Flags().Changed("json") && !viper.IsSet("json") {
			jsonOut = true
		}

		// Perform comprehensive test
		testResults := make(map[string]interface{})
		allPassed := true

		// Test 1: Environment validation
		envResponse := agent.ValidateAgentEnvironment()
		testResults["environment_validation"] = map[string]interface{}{
			"passed": envResponse.Success,
			"error":  envResponse.Error,
		}
		if !envResponse.Success {
			allPassed = false
		}

		// Test 2: Authentication check
		authStatus, err := auth.GetAuthStatus()
		testResults["authentication"] = map[string]interface{}{
			"passed": err == nil && authStatus.Authenticated,
			"method": authStatus.Method,
			"user":   authStatus.User,
			"error":  err,
		}
		if err != nil || !authStatus.Authenticated {
			allPassed = false
		}

		// Test 3: Actor configuration
		actorConfig := oauth.LoadActorFromEnvironment()
		testResults["actor_configuration"] = map[string]interface{}{
			"configured":         actorConfig.IsConfigured(),
			"default_actor":      actorConfig.DefaultActor,
			"default_avatar_url": actorConfig.DefaultAvatarURL,
		}

		// Create final response
		response := &agent.AgentResponse{
			Success:   allPassed,
			Data:      testResults,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Metadata: map[string]interface{}{
				"test_type": "comprehensive",
				"tests_run": len(testResults),
			},
		}

		if !allPassed {
			response.Error = &agent.AgentError{
				Code:    "TEST_FAILED",
				Message: "One or more agent tests failed",
				Suggestions: []string{
					"Check environment variable configuration",
					"Verify OAuth authentication is working",
					"Run 'linctl agent validate' for detailed validation",
				},
				Retryable: false,
			}
		}

		agent.ExitWithResponse(response, jsonOut)
	},
}

func init() {
	rootCmd.AddCommand(agentCmd)
	agentCmd.AddCommand(agentValidateCmd)
	agentCmd.AddCommand(agentStatusCmd)
	agentCmd.AddCommand(agentConfigCmd)
	agentCmd.AddCommand(agentTestCmd)
}
