package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/nicholls-inc/linctl/pkg/api"
	"github.com/nicholls-inc/linctl/pkg/auth"
	"github.com/nicholls-inc/linctl/pkg/output"
	"github.com/nicholls-inc/linctl/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// commentCmd represents the comment command
var commentCmd = &cobra.Command{
	Use:   "comment",
	Short: "Manage issue comments",
	Long: `Manage comments on Linear issues including listing and creating comments.

Examples:
  linctl comment list LIN-123        # List comments for an issue
  linctl comment create LIN-123 --body "This is fixed"  # Add a comment
  linctl comment create LIN-123 --body "Working on this" --actor "AI Agent"  # Add comment with actor attribution`,
}

var commentListCmd = &cobra.Command{
	Use:     "list ISSUE-ID",
	Aliases: []string{"ls"},
	Short:   "List comments for an issue",
	Long:    `List all comments for a specific issue.`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		issueID := args[0]

		// Get auth header
		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Create API client
		client := api.NewClient(authHeader)

		// Get limit
		limit, _ := cmd.Flags().GetInt("limit")

		// Get sort option
		sortBy, _ := cmd.Flags().GetString("sort")
		orderBy := ""
		if sortBy != "" {
			switch sortBy {
			case "created", "createdAt":
				orderBy = "createdAt"
			case "updated", "updatedAt":
				orderBy = "updatedAt"
			case "linear":
				// Use empty string for Linear's default sort
				orderBy = ""
			default:
				output.Error(fmt.Sprintf("Invalid sort option: %s. Valid options are: linear, created, updated", sortBy), plaintext, jsonOut)
				os.Exit(1)
			}
		}

		// Get comments
		comments, err := client.GetIssueComments(context.Background(), issueID, limit, "", orderBy)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to list comments: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Handle output
		if jsonOut {
			output.JSON(comments.Nodes)
		} else if plaintext {
			for i, comment := range comments.Nodes {
				if i > 0 {
					fmt.Println("---")
				}
				authorName := "Unknown"
				if comment.User != nil {
					authorName = comment.User.Name
				}
				fmt.Printf("Author: %s\n", authorName)
				fmt.Printf("Date: %s\n", comment.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("Comment:\n%s\n", comment.Body)
			}
		} else {
			// Rich display
			if len(comments.Nodes) == 0 {
				fmt.Printf("\n%s No comments on issue %s\n",
					color.New(color.FgYellow).Sprint("ℹ️"),
					color.New(color.FgCyan).Sprint(issueID))
				return
			}

			fmt.Printf("\n%s Comments on %s (%d)\n\n",
				color.New(color.FgCyan, color.Bold).Sprint("💬"),
				color.New(color.FgCyan).Sprint(issueID),
				len(comments.Nodes))

			for i, comment := range comments.Nodes {
				if i > 0 {
					fmt.Println(strings.Repeat("─", 50))
				}

				// Header with author and time
				timeAgo := formatTimeAgo(comment.CreatedAt)
				authorName := "Unknown"
				if comment.User != nil {
					authorName = comment.User.Name
				}
				fmt.Printf("%s %s %s\n",
					color.New(color.FgCyan, color.Bold).Sprint(authorName),
					color.New(color.FgWhite, color.Faint).Sprint("•"),
					color.New(color.FgWhite, color.Faint).Sprint(timeAgo))

				// Comment body
				fmt.Printf("\n%s\n\n", comment.Body)
			}
		}
	},
}

var commentCreateCmd = &cobra.Command{
	Use:     "create ISSUE-ID",
	Aliases: []string{"add", "new"},
	Short:   "Create a comment on an issue",
	Long:    `Add a new comment to a specific issue.`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		issueID := args[0]

		// Get auth header
		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Create API client
		client := api.NewClient(authHeader)

		// Get comment body and actor parameters
		body, _ := cmd.Flags().GetString("body")
		actor, _ := cmd.Flags().GetString("actor")
		avatarURL, _ := cmd.Flags().GetString("avatar-url")

		if body == "" {
			output.Error("Comment body is required (--body)", plaintext, jsonOut)
			os.Exit(1)
		}

		// Resolve actor parameters
		actorParams := utils.ResolveActorParams(actor, avatarURL)

		// Build input
		input := api.CommentCreateInput{
			IssueID:        issueID,
			Body:           body,
			CreateAsUser:   actorParams.ToCreateAsUser(),
			DisplayIconURL: actorParams.ToDisplayIconURL(),
		}

		// Create comment
		comment, err := client.CreateComment(context.Background(), input)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to create comment: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Handle output
		if jsonOut {
			output.JSON(comment)
		} else if plaintext {
			fmt.Printf("Created comment on %s\n", issueID)
			authorName := "Unknown"
			if comment.User != nil {
				authorName = comment.User.Name
			}
			fmt.Printf("Author: %s\n", authorName)
			fmt.Printf("Date: %s\n", comment.CreatedAt.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("%s Added comment to %s\n",
				color.New(color.FgGreen).Sprint("✓"),
				color.New(color.FgCyan, color.Bold).Sprint(issueID))
			fmt.Printf("\n%s\n", comment.Body)
		}
	},
}

// formatTimeAgo formats a time as a human-readable "time ago" string
func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if duration < 30*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else if duration < 365*24*time.Hour {
		months := int(duration.Hours() / (24 * 30))
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	} else {
		years := int(duration.Hours() / (24 * 365))
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

func init() {
	rootCmd.AddCommand(commentCmd)
	commentCmd.AddCommand(commentListCmd)
	commentCmd.AddCommand(commentCreateCmd)

	// List command flags
	commentListCmd.Flags().IntP("limit", "l", 50, "Maximum number of comments to return")
	commentListCmd.Flags().StringP("sort", "o", "linear", "Sort order: linear (default), created, updated")

	// Create command flags
	commentCreateCmd.Flags().StringP("body", "b", "", "Comment body (required)")
	commentCreateCmd.Flags().String("actor", "", "Actor name for attribution (uses LINEAR_DEFAULT_ACTOR if not specified)")
	commentCreateCmd.Flags().String("avatar-url", "", "Avatar URL for actor (uses LINEAR_DEFAULT_AVATAR_URL if not specified)")
	_ = commentCreateCmd.MarkFlagRequired("body")
}
