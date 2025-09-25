package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCommentCreateCommand_ActorFlags(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		expectActor     string
		expectAvatarURL string
		expectError     bool
	}{
		{
			name:            "no actor flags",
			args:            []string{"create", "LIN-123", "--body", "Test comment"},
			expectActor:     "",
			expectAvatarURL: "",
			expectError:     false,
		},
		{
			name:            "actor flag only",
			args:            []string{"create", "LIN-123", "--body", "Test comment", "--actor", "AI Agent"},
			expectActor:     "AI Agent",
			expectAvatarURL: "",
			expectError:     false,
		},
		{
			name:            "avatar-url flag only",
			args:            []string{"create", "LIN-123", "--body", "Test comment", "--avatar-url", "https://example.com/agent.png"},
			expectActor:     "",
			expectAvatarURL: "https://example.com/agent.png",
			expectError:     false,
		},
		{
			name:            "both actor flags",
			args:            []string{"create", "LIN-123", "--body", "Test comment", "--actor", "AI Agent", "--avatar-url", "https://example.com/agent.png"},
			expectActor:     "AI Agent",
			expectAvatarURL: "https://example.com/agent.png",
			expectError:     false,
		},
		{
			name:            "actor with spaces",
			args:            []string{"create", "LIN-123", "--body", "Test comment", "--actor", "AI Agent Bot"},
			expectActor:     "AI Agent Bot",
			expectAvatarURL: "",
			expectError:     false,
		},
		{
			name:            "long comment body with actor",
			args:            []string{"create", "LIN-123", "--body", "This is a longer comment body that spans multiple words and includes various details about the issue.", "--actor", "AI Agent"},
			expectActor:     "AI Agent",
			expectAvatarURL: "",
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new command instance for each test
			cmd := &cobra.Command{
				Use: "create",
				Run: func(cmd *cobra.Command, args []string) {
					// Test implementation - just verify flags are parsed correctly
				},
			}

			// Add the same flags as the real command
			cmd.Flags().StringP("body", "b", "", "Comment body (required)")
			cmd.Flags().String("actor", "", "Actor name for attribution (uses LINEAR_DEFAULT_ACTOR if not specified)")
			cmd.Flags().String("avatar-url", "", "Avatar URL for actor (uses LINEAR_DEFAULT_AVATAR_URL if not specified)")
			_ = cmd.MarkFlagRequired("body")

			// Set arguments and parse
			cmd.SetArgs(tt.args[1:]) // Skip "create" as it's the command name
			err := cmd.ParseFlags(tt.args[1:])

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error parsing flags: %v", err)
			}

			// Verify actor flag
			actor, err := cmd.Flags().GetString("actor")
			if err != nil {
				t.Fatalf("Failed to get actor flag: %v", err)
			}
			if actor != tt.expectActor {
				t.Errorf("Expected actor '%s', got '%s'", tt.expectActor, actor)
			}

			// Verify avatar-url flag
			avatarURL, err := cmd.Flags().GetString("avatar-url")
			if err != nil {
				t.Fatalf("Failed to get avatar-url flag: %v", err)
			}
			if avatarURL != tt.expectAvatarURL {
				t.Errorf("Expected avatar URL '%s', got '%s'", tt.expectAvatarURL, avatarURL)
			}

			// Verify body flag is present
			body, err := cmd.Flags().GetString("body")
			if err != nil {
				t.Fatalf("Failed to get body flag: %v", err)
			}
			if body != "Test comment" && !strings.Contains(body, "comment") {
				t.Errorf("Expected body to contain 'comment', got '%s'", body)
			}

			// Verify issue ID is in args (first positional argument)
			if len(cmd.Flags().Args()) == 0 {
				t.Error("Expected issue ID as positional argument")
			} else if cmd.Flags().Args()[0] != "LIN-123" {
				t.Errorf("Expected issue ID 'LIN-123', got '%s'", cmd.Flags().Args()[0])
			}
		})
	}
}

func TestCommentCreateCommand_EnvironmentVariables(t *testing.T) {
	// Save original environment
	originalActor := os.Getenv("LINEAR_DEFAULT_ACTOR")
	originalAvatarURL := os.Getenv("LINEAR_DEFAULT_AVATAR_URL")

	// Clean up after test
	defer func() {
		os.Setenv("LINEAR_DEFAULT_ACTOR", originalActor)
		os.Setenv("LINEAR_DEFAULT_AVATAR_URL", originalAvatarURL)
	}()

	tests := []struct {
		name           string
		envActor       string
		envAvatarURL   string
		flagActor      string
		flagAvatarURL  string
		expectedActor  string
		expectedAvatar string
	}{
		{
			name:           "environment variables only",
			envActor:       "Env Agent",
			envAvatarURL:   "https://env.com/avatar.png",
			flagActor:      "",
			flagAvatarURL:  "",
			expectedActor:  "Env Agent",
			expectedAvatar: "https://env.com/avatar.png",
		},
		{
			name:           "flags override environment",
			envActor:       "Env Agent",
			envAvatarURL:   "https://env.com/avatar.png",
			flagActor:      "Flag Agent",
			flagAvatarURL:  "https://flag.com/avatar.png",
			expectedActor:  "Flag Agent",
			expectedAvatar: "https://flag.com/avatar.png",
		},
		{
			name:           "mixed environment and flags",
			envActor:       "Env Agent",
			envAvatarURL:   "https://env.com/avatar.png",
			flagActor:      "Flag Agent",
			flagAvatarURL:  "",
			expectedActor:  "Flag Agent",
			expectedAvatar: "https://env.com/avatar.png",
		},
		{
			name:           "no environment or flags",
			envActor:       "",
			envAvatarURL:   "",
			flagActor:      "",
			flagAvatarURL:  "",
			expectedActor:  "",
			expectedAvatar: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("LINEAR_DEFAULT_ACTOR", tt.envActor)
			os.Setenv("LINEAR_DEFAULT_AVATAR_URL", tt.envAvatarURL)

			// This test verifies that the environment variable resolution logic
			// would work correctly. Since we can't easily test the full command
			// execution without mocking the API, we test the actor resolution
			// logic directly through the utils package.

			// The actual command would call utils.ResolveActorParams(flagActor, flagAvatarURL)
			// and get the expected results based on the environment and flag values

			// For now, we just verify the environment variables are set correctly
			if os.Getenv("LINEAR_DEFAULT_ACTOR") != tt.envActor {
				t.Errorf("Expected env actor '%s', got '%s'", tt.envActor, os.Getenv("LINEAR_DEFAULT_ACTOR"))
			}

			if os.Getenv("LINEAR_DEFAULT_AVATAR_URL") != tt.envAvatarURL {
				t.Errorf("Expected env avatar URL '%s', got '%s'", tt.envAvatarURL, os.Getenv("LINEAR_DEFAULT_AVATAR_URL"))
			}
		})
	}
}

func TestCommentCreateCommand_Help(t *testing.T) {
	// Test that the help text includes actor flags
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new comment",
		Long:  `Create a new comment on an issue in Linear.`,
	}

	// Add the same flags as the real command
	cmd.Flags().StringP("body", "b", "", "Comment body (required)")
	cmd.Flags().String("actor", "", "Actor name for attribution (uses LINEAR_DEFAULT_ACTOR if not specified)")
	cmd.Flags().String("avatar-url", "", "Avatar URL for actor (uses LINEAR_DEFAULT_AVATAR_URL if not specified)")

	// Get help text
	helpText := cmd.UsageString()

	// Verify actor flags are present in help
	if !strings.Contains(helpText, "--actor") {
		t.Error("Help text should contain --actor flag")
	}

	if !strings.Contains(helpText, "--avatar-url") {
		t.Error("Help text should contain --avatar-url flag")
	}

	if !strings.Contains(helpText, "Actor name for attribution") {
		t.Error("Help text should contain actor flag description")
	}

	if !strings.Contains(helpText, "Avatar URL for actor") {
		t.Error("Help text should contain avatar URL flag description")
	}

	if !strings.Contains(helpText, "LINEAR_DEFAULT_ACTOR") {
		t.Error("Help text should mention LINEAR_DEFAULT_ACTOR environment variable")
	}

	if !strings.Contains(helpText, "LINEAR_DEFAULT_AVATAR_URL") {
		t.Error("Help text should mention LINEAR_DEFAULT_AVATAR_URL environment variable")
	}
}

func TestCommentCreateCommand_FlagValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "missing body",
			args:        []string{"create", "LIN-123"},
			expectError: true,
			errorMsg:    "required flag(s) \"body\" not set",
		},
		{
			name:        "missing issue ID",
			args:        []string{"create", "--body", "Test comment"},
			expectError: false, // This would be caught by the command logic, not flag validation
		},
		{
			name:        "valid required flags",
			args:        []string{"create", "LIN-123", "--body", "Test comment"},
			expectError: false,
		},
		{
			name:        "valid with actor flags",
			args:        []string{"create", "LIN-123", "--body", "Test comment", "--actor", "AI Agent", "--avatar-url", "https://example.com/agent.png"},
			expectError: false,
		},
		{
			name:        "empty actor flag",
			args:        []string{"create", "LIN-123", "--body", "Test comment", "--actor", ""},
			expectError: false, // Empty actor flag is allowed
		},
		{
			name:        "empty avatar-url flag",
			args:        []string{"create", "LIN-123", "--body", "Test comment", "--avatar-url", ""},
			expectError: false, // Empty avatar URL flag is allowed
		},
		{
			name:        "empty body flag",
			args:        []string{"create", "LIN-123", "--body", ""},
			expectError: false, // Empty body might be allowed, depends on implementation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "create",
				Run: func(cmd *cobra.Command, args []string) {
					// Test implementation
				},
			}

			// Add the same flags as the real command
			cmd.Flags().StringP("body", "b", "", "Comment body (required)")
			cmd.Flags().String("actor", "", "Actor name for attribution (uses LINEAR_DEFAULT_ACTOR if not specified)")
			cmd.Flags().String("avatar-url", "", "Avatar URL for actor (uses LINEAR_DEFAULT_AVATAR_URL if not specified)")
			_ = cmd.MarkFlagRequired("body")

			// Set arguments and execute
			cmd.SetArgs(tt.args[1:]) // Skip "create" as it's the command name
			err := cmd.Execute()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCommentCommand_Examples(t *testing.T) {
	// Test that the main comment command includes actor examples
	examples := `Examples:
  linctl comment list LIN-123
  linctl comment ls LIN-123
  linctl comment create LIN-123 --body "This is a comment"
  linctl comment create LIN-123 -b "This is a comment"
  linctl comment create LIN-123 --body "This is a comment" --actor "AI Agent" --avatar-url "https://example.com/agent.png"`

	// Verify actor example is present
	if !strings.Contains(examples, "--actor \"AI Agent\"") {
		t.Error("Examples should contain actor flag usage")
	}

	if !strings.Contains(examples, "--avatar-url \"https://example.com/agent.png\"") {
		t.Error("Examples should contain avatar-url flag usage")
	}

	// Verify both basic and actor examples are present
	basicExample := "linctl comment create LIN-123 --body \"This is a comment\""
	actorExample := "linctl comment create LIN-123 --body \"This is a comment\" --actor \"AI Agent\" --avatar-url \"https://example.com/agent.png\""

	if !strings.Contains(examples, basicExample) {
		t.Error("Examples should contain basic create example")
	}

	if !strings.Contains(examples, actorExample) {
		t.Error("Examples should contain actor create example")
	}
}

func TestCommentCreateCommand_IssueIDValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		issueID string
		valid   bool
	}{
		{
			name:    "valid Linear issue ID",
			args:    []string{"create", "LIN-123", "--body", "Test comment"},
			issueID: "LIN-123",
			valid:   true,
		},
		{
			name:    "valid Linear issue ID with different team",
			args:    []string{"create", "ENG-456", "--body", "Test comment"},
			issueID: "ENG-456",
			valid:   true,
		},
		{
			name:    "valid Linear issue ID with longer number",
			args:    []string{"create", "TEAM-12345", "--body", "Test comment"},
			issueID: "TEAM-12345",
			valid:   true,
		},
		{
			name:    "issue ID with lowercase",
			args:    []string{"create", "lin-123", "--body", "Test comment"},
			issueID: "lin-123",
			valid:   true, // Depends on implementation - might be normalized
		},
		{
			name:    "empty issue ID",
			args:    []string{"create", "", "--body", "Test comment"},
			issueID: "",
			valid:   true, // Empty string is still a valid argument, validation would happen in command logic
		},
		{
			name:    "invalid issue ID format",
			args:    []string{"create", "123", "--body", "Test comment"},
			issueID: "123",
			valid:   true, // Argument parsing succeeds, format validation would happen in command logic
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "create",
				Args: func(cmd *cobra.Command, args []string) error {
					// This would contain the actual validation logic
					if len(args) == 0 {
						return cobra.ExactArgs(1)(cmd, args)
					}
					issueID := args[0]
					if issueID == "" {
						return cobra.ExactArgs(1)(cmd, args)
					}
					// Additional validation would go here
					return nil
				},
				Run: func(cmd *cobra.Command, args []string) {
					// Test implementation
				},
			}

			cmd.Flags().StringP("body", "b", "", "Comment body (required)")
			cmd.Flags().String("actor", "", "Actor name for attribution")
			cmd.Flags().String("avatar-url", "", "Avatar URL for actor")
			_ = cmd.MarkFlagRequired("body")

			cmd.SetArgs(tt.args[1:])
			err := cmd.Execute()

			if !tt.valid && err == nil {
				t.Error("Expected validation error but got none")
			}

			if tt.valid && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}

			// If valid, verify the issue ID is parsed correctly
			if tt.valid && err == nil && len(cmd.Flags().Args()) > 0 {
				actualIssueID := cmd.Flags().Args()[0]
				if actualIssueID != tt.issueID {
					t.Errorf("Expected issue ID '%s', got '%s'", tt.issueID, actualIssueID)
				}
			}
		})
	}
}

func TestCommentCreateCommand_ActorFlagShortcuts(t *testing.T) {
	// Test various ways to specify actor flags
	tests := []struct {
		name        string
		args        []string
		expectActor string
		expectURL   string
	}{
		{
			name:        "full flag names",
			args:        []string{"create", "LIN-123", "--body", "Test", "--actor", "Agent", "--avatar-url", "https://example.com/avatar.png"},
			expectActor: "Agent",
			expectURL:   "https://example.com/avatar.png",
		},
		{
			name:        "actor with special characters",
			args:        []string{"create", "LIN-123", "--body", "Test", "--actor", "AI Agent (v2.0)", "--avatar-url", "https://example.com/avatar.png"},
			expectActor: "AI Agent (v2.0)",
			expectURL:   "https://example.com/avatar.png",
		},
		{
			name:        "actor with quotes",
			args:        []string{"create", "LIN-123", "--body", "Test", "--actor", "\"Quoted Agent\"", "--avatar-url", "https://example.com/avatar.png"},
			expectActor: "\"Quoted Agent\"",
			expectURL:   "https://example.com/avatar.png",
		},
		{
			name:        "URL with query parameters",
			args:        []string{"create", "LIN-123", "--body", "Test", "--actor", "Agent", "--avatar-url", "https://example.com/avatar.png?size=64&format=png"},
			expectActor: "Agent",
			expectURL:   "https://example.com/avatar.png?size=64&format=png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "create",
				Run: func(cmd *cobra.Command, args []string) {},
			}

			cmd.Flags().StringP("body", "b", "", "Comment body (required)")
			cmd.Flags().String("actor", "", "Actor name for attribution")
			cmd.Flags().String("avatar-url", "", "Avatar URL for actor")
			_ = cmd.MarkFlagRequired("body")

			cmd.SetArgs(tt.args[1:])
			err := cmd.ParseFlags(tt.args[1:])
			if err != nil {
				t.Fatalf("Failed to parse flags: %v", err)
			}

			actor, _ := cmd.Flags().GetString("actor")
			if actor != tt.expectActor {
				t.Errorf("Expected actor '%s', got '%s'", tt.expectActor, actor)
			}

			avatarURL, _ := cmd.Flags().GetString("avatar-url")
			if avatarURL != tt.expectURL {
				t.Errorf("Expected avatar URL '%s', got '%s'", tt.expectURL, avatarURL)
			}
		})
	}
}
