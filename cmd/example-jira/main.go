package main

import (
	"fmt"
	"log"
	"os"

	"github.com/matst80/go-jira/pkg/jira"
)

type jiraClient interface {
	ListBoardIssues(boardID int, jql string) (*jira.SearchResult, error)
	GetIssue(issueID string) (*jira.Issue, error)
	GetTransitions(issueID string) (*jira.TransitionsResult, error)
	FindUsers(query string) ([]jira.User, error)
	DownloadAttachment(attachment jira.Attachment) ([]byte, error)
	AddComment(issueID string, comment string) error
	AssignIssueByAccountID(issueID string, accountID string) error
	TransitionIssue(issueID string, transitionID string) error
}

func main() {
	token := os.Getenv("ATLASSIAN_KEY")
	if token == "" {
		log.Fatal("ATLASSIAN_KEY environment variable is required")
	}

	// The user request targets this specific Jira domain and board.
	baseURL := "https://elkjop.atlassian.net"
	boardID := 773

	logFile, err := os.Create("jira-response.log")
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	defer logFile.Close()

	client := jira.NewClient(baseURL, token, jira.WithDebugWriter(logFile))
	if email := os.Getenv("ATLASSIAN_EMAIL"); email != "" {
		client = jira.NewClient(baseURL, token, jira.WithBasicAuth(email), jira.WithDebugWriter(logFile))
	} else if prefix := os.Getenv("ATLASSIAN_AUTH_PREFIX"); prefix != "" {
		client = jira.NewClient(baseURL, token, jira.WithAuthPrefix(prefix), jira.WithDebugWriter(logFile))
	}

	if err := run(client, boardID); err != nil {
		log.Fatal(err)
	}
}

func run(client jiraClient, boardID int) error {
	jql := "labels = ai-task AND statusCategory != Done ORDER BY created DESC"
	issues, err := client.ListBoardIssues(boardID, jql)
	if err != nil {
		return fmt.Errorf("failed to fetch backlog tickets: %w", err)
	}

	users, err := client.FindUsers("mats.tornberg@elkjop.no")
	for _, user := range users {
		log.Printf("User %s, id:%s", user.DisplayName, user.AccountId)
	}

	fmt.Printf("Successfully fetched issues for tag ai-task!\n")
	fmt.Printf("Total Issues: %d\n", issues.Total)
	fmt.Printf("Showing up to %d items (startAt: %d)\n", issues.MaxResults, issues.StartAt)
	fmt.Println("--------------------------------------------------")

	for _, issue := range issues.Issues {
		assignee := "Unassigned"
		if issue.Fields.Assignee != nil {
			assignee = issue.Fields.Assignee.DisplayName
		}

		fmt.Printf("[%s] %s\n", issue.Key, issue.Fields.Summary)
		fmt.Printf("  └─ Status:   %s\n", issue.Fields.Status.Name)
		fmt.Printf("  └─ Type:     %s\n", issue.Fields.IssueType.Name)
		fmt.Printf("  └─ Assignee: %s\n", assignee)
		fmt.Printf("  └─ Created: %s\n", issue.Fields.Created)
		fmt.Printf("  └─ Updated: %s\n", issue.Fields.Updated)
		fmt.Printf("  └─ Description length: %d\n", len(issue.Fields.Description))
		fmt.Printf("  └─ Comments: %d\n", issue.Fields.Comment.Total)
		fmt.Printf("  └─ Attachments: %d\n", len(issue.Fields.Attachment))
		for _, att := range issue.Fields.Attachment {
			fmt.Printf("     - %s (%d bytes, ID: %s)\n", att.Filename, att.Size, att.ID)
		}
		fmt.Println()
	}

	var downloadedAttachments int

	// Example: Fetch a specific issue to see its detail and optionally mutate it.
	if len(issues.Issues) > 0 {
		for _, issue := range issues.Issues {

			fmt.Printf("Fetching full detail for %s...\n", issue.Key)
			detail, err := client.GetIssue(issue.Key)
			if err != nil {
				log.Printf("Failed to fetch issue detail: %v", err)
			} else {
				fmt.Printf("Detail for %s: %d attachments\n", detail.Key, len(detail.Fields.Attachment))

				transitions, err := client.GetTransitions(detail.Key)
				if err != nil {
					fmt.Printf("Failed to fetch transitions for %s: %v\n", detail.Key, err)
				} else {
					fmt.Printf("Available transitions for %s on board %d (NEP):\n", detail.Key, boardID)
					if len(transitions.Transitions) == 0 {
						fmt.Println("  (no transitions available)")
					} else {
						for _, transition := range transitions.Transitions {
							targetStatus := ""
							if transition.To != nil && transition.To.Name != "" {
								targetStatus = fmt.Sprintf(" -> %s", transition.To.Name)
							}
							fmt.Printf("  - %s: %s%s\n", transition.ID, transition.Name, targetStatus)
						}
					}
				}

				fmt.Printf("Adding comment to %s...\n", detail.Key)
				if err := client.AddComment(detail.Key, "test comment from automation"); err != nil {
					fmt.Printf("Failed to add comment: %v\n", err)
				} else {
					fmt.Printf("Successfully added comment to %s\n", detail.Key)
				}

				assigneeQuery := "mats.tornberg@elkjop.no"
				fmt.Printf("Looking up Jira user for %s...\n", assigneeQuery)
				users, err := client.FindUsers(assigneeQuery)
				if err != nil {
					fmt.Printf("Failed to look up Jira user: %v\n", err)
				} else if len(users) == 0 {
					fmt.Printf("No Jira users found for %s\n", assigneeQuery)
				} else {
					fmt.Println("Matching Jira users:")
					for _, user := range users {
						fmt.Printf("  - %s | accountId=%s | email=%s\n", user.DisplayName, user.AccountId, user.EmailAddress)
					}

					accountID := users[0].AccountId
					fmt.Printf("Assigning %s to accountId %s...\n", detail.Key, accountID)
					if err := client.AssignIssueByAccountID(detail.Key, accountID); err != nil {
						fmt.Printf("Failed to assign issue: %v\n", err)
					} else {
						fmt.Printf("Successfully assigned %s\n", detail.Key)
					}
				}

				if transitionID := os.Getenv("EXAMPLE_TRANSITION_ID"); transitionID != "" {
					fmt.Printf("Transitioning %s with transition %s...\n", detail.Key, transitionID)
					if err := client.TransitionIssue(detail.Key, transitionID); err != nil {
						fmt.Printf("Failed to transition issue: %v\n", err)
					} else {
						fmt.Printf("Successfully transitioned %s\n", detail.Key)
					}
				}

				if len(detail.Fields.Attachment) > 0 {
					for _, att := range detail.Fields.Attachment {

						fmt.Printf("Downloading first attachment: %s...\n", att.Filename)
						if downloadedAttachments < 5 {
							downloadedAttachments++
							content, err := client.DownloadAttachment(att)
							if err == nil {
								fmt.Printf("Downloaded %d bytes\n", len(content))
								SaveImage(content, att.Filename)
							} else {
								fmt.Printf("Failed to download attachment: %v\n", err)
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func SaveImage(content []byte, filename string) {
	err := os.WriteFile(filename, content, 0644)
	if err != nil {
		fmt.Printf("Failed to save image %s: %v\n", filename, err)
		return
	}
	fmt.Printf("Successfully saved image to %s\n", filename)
}
